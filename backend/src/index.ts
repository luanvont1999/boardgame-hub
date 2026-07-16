import { Hono } from "hono";
import { cors } from "hono/cors";
import { serve } from "@hono/node-server";
import fs from "fs";
import path from "path";
import jwt from "jsonwebtoken";
import dotenv from "dotenv";

import { authMiddleware, FirebaseUser } from "./middleware/auth.js";
import { meetupStore, Meetup } from "./store/meetups.js";

dotenv.config();

type Variables = {
  firebase_user: FirebaseUser;
};

const app = new Hono<{ Variables: Variables }>();

// Enable CORS for all routes
app.use(
  "*",
  cors({
    origin: "*",
    allowMethods: ["GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"],
    allowHeaders: ["Content-Type", "Authorization"],
  })
);

interface ServiceAccount {
  type: string;
  project_id: string;
  private_key: string;
  client_email: string;
  token_uri: string;
}

let cachedAccessToken = "";
let tokenExpiration = 0;
let activeProjectId = "";

// Helper to load service account and get access token
async function getAccessToken(): Promise<{ accessToken: string; projectId: string }> {
  const now = Date.now();
  if (cachedAccessToken && now < tokenExpiration) {
    return { accessToken: cachedAccessToken, projectId: activeProjectId };
  }

  let sa: ServiceAccount;
  const saJSON = process.env.FIREBASE_SERVICE_ACCOUNT_JSON;

  if (saJSON) {
    try {
      sa = JSON.parse(saJSON);
    } catch (err: any) {
      throw new Error(`Lỗi parse service account từ env: ${err.message}`);
    }
  } else {
    let saPath = process.env.FIREBASE_SERVICE_ACCOUNT_JSON_PATH || "firebase-service-account.json";
    if (!path.isAbsolute(saPath)) {
      // Look in workspace root
      saPath = path.resolve(process.cwd(), "..", saPath);
    }

    if (!fs.existsSync(saPath)) {
      // Try local fallback
      saPath = path.resolve(process.cwd(), "firebase-service-account.json");
    }

    try {
      const fileContent = fs.readFileSync(saPath, "utf-8");
      sa = JSON.parse(fileContent);
    } catch (err: any) {
      throw new Error(`Không thể đọc file service account tại ${saPath}: ${err.message}`);
    }
  }

  // Create assertion token signed with private key (RS256)
  const assertion = jwt.sign(
    {
      iss: sa.client_email,
      sub: sa.client_email,
      aud: sa.token_uri,
      scope: "https://www.googleapis.com/auth/cloud-platform",
    },
    sa.private_key,
    {
      algorithm: "RS256",
      expiresIn: "1h",
    }
  );

  // Request Access Token from Google Token URI
  const res = await fetch(sa.token_uri, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: new URLSearchParams({
      grant_type: "urn:ietf:params:oauth:grant-type:jwt-bearer",
      assertion: assertion,
    }),
  });

  if (!res.ok) {
    const text = await res.text();
    throw new Error(`Lỗi Google OAuth API: ${res.status} - ${text}`);
  }

  const data = (await res.json()) as { access_token: string; expires_in?: number };
  cachedAccessToken = data.access_token;
  // Cache for slightly less than 1 hour (e.g. 50 mins)
  tokenExpiration = now + (data.expires_in || 3600) * 1000 - 5 * 60 * 1000;
  activeProjectId = sa.project_id;

  return { accessToken: cachedAccessToken, projectId: activeProjectId };
}

// FCM Notification Helper
async function sendFCMNotification(
  token: string,
  title: string,
  body: string,
  clickAction: string
): Promise<void> {
  const { accessToken, projectId } = await getAccessToken();
  const url = `https://fcm.googleapis.com/v1/projects/${projectId}/messages:send`;

  const payload = {
    message: {
      token: token,
      data: {
        title,
        body,
        clickAction,
      },
      webpush: {
        headers: {
          Urgency: "high",
        },
        fcm_options: {
          link: clickAction,
        },
      },
      apns: {
        headers: {
          "apns-priority": "10",
        },
        payload: {
          aps: {
            alert: {
              title,
              body,
            },
            sound: "default",
            "mutable-content": 1,
            "content-available": 1,
          },
        },
      },
    },
  };

  const res = await fetch(url, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${accessToken}`,
    },
    body: JSON.stringify(payload),
  });

  if (!res.ok) {
    const text = await res.text();
    throw new Error(`FCM V1 API error ${res.status}: ${text}`);
  }
}

// ── FIRESTORE REST PAYLOAD PARSERS & HELPERS ──────────────────────────────

interface FirestoreField {
  stringValue?: string;
  integerValue?: string;
  arrayValue?: { values: FirestoreField[] };
  mapValue?: { fields: Record<string, FirestoreField> };
}

interface FirestoreDocument {
  name?: string;
  fields: Record<string, FirestoreField>;
}

interface MeetupData {
  id: string;
  title: string;
  game: string;
  hostName: string;
  hostUID: string;
  hostFcmToken: string;
  playersCount: number;
  playersNeeded: number;
  time: string;
  color: string;
  pendingUids: string[];
  approvedPendingUids: string[];
  approvedUids: string[];
  userFcmTokens: Record<string, string>;
}

function parseFirestoreDocument(doc: FirestoreDocument): MeetupData {
  const fields = doc.fields || {};
  const m: MeetupData = {
    id: "",
    title: "",
    game: "",
    hostName: "",
    hostUID: "",
    hostFcmToken: "",
    playersCount: 1,
    playersNeeded: 4,
    time: "",
    color: "",
    pendingUids: [],
    approvedPendingUids: [],
    approvedUids: [],
    userFcmTokens: {},
  };

  if (doc.name) {
    const parts = doc.name.split("/");
    m.id = parts[parts.length - 1];
  }

  m.title = fields.title?.stringValue || "";
  m.game = fields.game?.stringValue || "";
  m.hostName = fields.hostName?.stringValue || fields.host_name?.stringValue || "";
  m.hostUID = fields.hostUid?.stringValue || fields.host_uid?.stringValue || "";
  m.hostFcmToken = fields.hostFcmToken?.stringValue || "";
  m.playersCount = parseInt(fields.playersCount?.integerValue || fields.players_count?.integerValue || "1");
  m.playersNeeded = parseInt(fields.playersNeeded?.integerValue || fields.players_needed?.integerValue || "4");
  m.time = fields.time?.stringValue || "";
  m.color = fields.color?.stringValue || "";

  if (fields.pendingUids?.arrayValue?.values) {
    m.pendingUids = fields.pendingUids.arrayValue.values
      .map((v) => v.stringValue || "")
      .filter(Boolean);
  }
  if (fields.approvedPendingUids?.arrayValue?.values) {
    m.approvedPendingUids = fields.approvedPendingUids.arrayValue.values
      .map((v) => v.stringValue || "")
      .filter(Boolean);
  }
  if (fields.approvedUids?.arrayValue?.values) {
    m.approvedUids = fields.approvedUids.arrayValue.values
      .map((v) => v.stringValue || "")
      .filter(Boolean);
  }

  if (fields.userFcmTokens?.mapValue?.fields) {
    const tokenMap = fields.userFcmTokens.mapValue.fields;
    for (const key of Object.keys(tokenMap)) {
      if (tokenMap[key]?.stringValue) {
        m.userFcmTokens[key] = tokenMap[key].stringValue!;
      }
    }
  }

  return m;
}

function buildFirestoreDocument(m: MeetupData): FirestoreDocument {
  const fields: Record<string, FirestoreField> = {
    title: { stringValue: m.title },
    game: { stringValue: m.game },
    hostName: { stringValue: m.hostName },
    hostUid: { stringValue: m.hostUID },
    hostFcmToken: { stringValue: m.hostFcmToken },
    playersCount: { integerValue: m.playersCount.toString() },
    playersNeeded: { integerValue: m.playersNeeded.toString() },
    time: { stringValue: m.time },
    color: { stringValue: m.color },
    pendingUids: {
      arrayValue: {
        values: m.pendingUids.map((uid) => ({ stringValue: uid })),
      },
    },
    approvedPendingUids: {
      arrayValue: {
        values: m.approvedPendingUids.map((uid) => ({ stringValue: uid })),
      },
    },
    approvedUids: {
      arrayValue: {
        values: m.approvedUids.map((uid) => ({ stringValue: uid })),
      },
    },
  };

  const tokenFields: Record<string, FirestoreField> = {};
  for (const key of Object.keys(m.userFcmTokens)) {
    tokenFields[key] = { stringValue: m.userFcmTokens[key] };
  }
  fields.userFcmTokens = { mapValue: { fields: tokenFields } };

  return { fields };
}

async function getFirestoreMeetup(meetupId: string): Promise<MeetupData> {
  const { accessToken, projectId } = await getAccessToken();
  const url = `https://firestore.googleapis.com/v1/projects/${projectId}/databases/(default)/documents/meetups/${meetupId}`;

  const res = await fetch(url, {
    headers: { Authorization: `Bearer ${accessToken}` },
  });

  if (!res.ok) {
    const text = await res.text();
    throw new Error(`Firestore read failed status ${res.status}: ${text}`);
  }

  const doc = (await res.json()) as FirestoreDocument;
  return parseFirestoreDocument(doc);
}

async function updateFirestoreMeetup(m: MeetupData, updateFields: string[]): Promise<void> {
  const { accessToken, projectId } = await getAccessToken();
  const fullDoc = buildFirestoreDocument(m);

  const filteredFields: Record<string, FirestoreField> = {};
  for (const field of updateFields) {
    if (fullDoc.fields[field] !== undefined) {
      filteredFields[field] = fullDoc.fields[field];
    }
  }

  const doc = { fields: filteredFields };
  const queryParams = new URLSearchParams();
  for (const field of updateFields) {
    queryParams.append("updateMask.fieldPaths", field);
  }

  const url = `https://firestore.googleapis.com/v1/projects/${projectId}/databases/(default)/documents/meetups/${m.id}?${queryParams.toString()}`;

  const res = await fetch(url, {
    method: "PATCH",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${accessToken}`,
    },
    body: JSON.stringify(doc),
  });

  if (!res.ok) {
    const text = await res.text();
    throw new Error(`Firestore patch failed status ${res.status}: ${text}`);
  }
}

async function setFirestoreRequest(
  meetupId: string,
  userUid: string,
  userName: string,
  status: string
): Promise<void> {
  const { accessToken, projectId } = await getAccessToken();
  const url = `https://firestore.googleapis.com/v1/projects/${projectId}/databases/(default)/documents/meetups/${meetupId}/requests/${userUid}`;

  const doc = {
    fields: {
      uid: { stringValue: userUid },
      name: { stringValue: userName },
      status: { stringValue: status },
    },
  };

  const res = await fetch(url, {
    method: "PATCH",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${accessToken}`,
    },
    body: JSON.stringify(doc),
  });

  if (!res.ok) {
    const text = await res.text();
    throw new Error(`Firestore set request failed status ${res.status}: ${text}`);
  }
}

async function updateFirestoreRequestStatus(
  meetupId: string,
  userUid: string,
  status: string
): Promise<void> {
  const { accessToken, projectId } = await getAccessToken();
  const url = `https://firestore.googleapis.com/v1/projects/${projectId}/databases/(default)/documents/meetups/${meetupId}/requests/${userUid}?updateMask.fieldPaths=status`;

  const doc = {
    fields: {
      status: { stringValue: status },
    },
  };

  const res = await fetch(url, {
    method: "PATCH",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${accessToken}`,
    },
    body: JSON.stringify(doc),
  });

  if (!res.ok) {
    const text = await res.text();
    throw new Error(`Firestore update request status failed: ${text}`);
  }
}

async function deleteFirestoreRequest(meetupId: string, userUid: string): Promise<void> {
  const { accessToken, projectId } = await getAccessToken();
  const url = `https://firestore.googleapis.com/v1/projects/${projectId}/databases/(default)/documents/meetups/${meetupId}/requests/${userUid}`;

  const res = await fetch(url, {
    method: "DELETE",
    headers: { Authorization: `Bearer ${accessToken}` },
  });

  if (!res.ok && res.status !== 404) {
    const text = await res.text();
    throw new Error(`Firestore delete request failed status ${res.status}: ${text}`);
  }
}

// ── ROUTE ENDPOINTS ────────────────────────────────────────────────────────

// Health Check API
app.get("/api/health", (c) => {
  return c.json({
    status: "OK",
    message: "Boardgame Luna API Node/Hono is running smoothly",
  });
});

// Authenticated User Profile API
app.get("/api/profile", authMiddleware, (c) => {
  const user = c.get("firebase_user") as FirebaseUser;
  return c.json({
    message: "Kết nối API xác thực thành công!",
    user,
  });
});

// Fallback GET Meetups API
app.get("/api/meetups", (c) => {
  return c.json(meetupStore.getAll());
});

// Authenticated POST Create Meetup API
app.post("/api/meetups/create", authMiddleware, async (c) => {
  const user = c.get("firebase_user") as FirebaseUser;
  let body: any;
  try {
    body = await c.req.json();
  } catch {
    c.status(400);
    return c.json({ error: "Invalid JSON body" });
  }

  const { title, game, lat, lng, time } = body;
  if (!title || !game || !lat || !lng || !time) {
    c.status(400);
    return c.json({ error: "Missing required fields (title, game, lat, lng, time)" });
  }

  const colors = ["#bca0f5", "#ffa4b2", "#ffe869", "#ffb875", "#9ee3b2", "#a4f0fd"];
  const color = colors[Math.floor(Math.random() * colors.length)];
  const hostName = user.name || user.email || "Ẩn danh";

  const newMeetup: Meetup = {
    id: Math.floor(Math.random() * 1000000).toString(),
    title,
    game,
    host_name: hostName,
    host_uid: user.uid,
    lat: parseFloat(lat),
    lng: parseFloat(lng),
    players_count: 1,
    players_needed: 4,
    time,
    color,
  };

  meetupStore.add(newMeetup);
  console.log(`[Meetups] Created new meetup: ${newMeetup.title} by ${newMeetup.host_name}`);

  c.status(201);
  return c.json(newMeetup);
});

// Proxy Send Notification API
app.post("/api/send-notification", async (c) => {
  let body: any;
  try {
    body = await c.req.json();
  } catch {
    c.status(400);
    return c.json({ error: "Invalid JSON body" });
  }

  const { fcmToken, fcmTokens, title, body: bodyText, clickAction } = body;
  if (!title || !bodyText) {
    c.status(400);
    return c.json({ error: "Missing required fields (title, body)" });
  }

  const tokens: string[] = [];
  if (fcmToken) tokens.push(fcmToken);
  if (Array.isArray(fcmTokens)) tokens.push(...fcmTokens);

  if (tokens.length === 0) {
    c.status(400);
    return c.json({ error: "At least one fcmToken or fcmTokens must be provided" });
  }

  console.log(`[FCM] Sending push notification to ${tokens.length} devices...`);

  const sendErrors: string[] = [];
  for (const token of tokens) {
    const tokenHint = token.length > 15 ? token.substring(0, 15) + "..." : token;
    try {
      await sendFCMNotification(token, title, bodyText, clickAction || "/");
    } catch (err: any) {
      console.error(`[FCM ERROR] Failed sending to ${tokenHint}:`, err);
      sendErrors.push(`${tokenHint}: ${err.message}`);
    }
  }

  if (sendErrors.length > 0 && sendErrors.length === tokens.length) {
    c.status(500);
    return c.json({
      success: false,
      warning: "Tất cả các lượt gửi đều thất bại. Vui lòng kiểm tra lại cấu hình hoặc token.",
      errors: sendErrors,
    });
  }

  return c.json({
    success: true,
    message: `Đã gửi thông báo đẩy thành công tới ${tokens.length - sendErrors.length}/${tokens.length} thiết bị!`,
    errors: sendErrors,
  });
});

// Join Meetup API
app.post("/api/meetups/join", async (c) => {
  let body: any;
  try {
    body = await c.req.json();
  } catch {
    c.status(400);
    return c.json({ error: "Invalid JSON body" });
  }

  const { meetupId, userUid, userName, fcmToken } = body;
  if (!meetupId || !userUid || !userName) {
    c.status(400);
    return c.json({ error: "Missing required fields" });
  }

  try {
    // 1. Create pending request record in Firestore
    await setFirestoreRequest(meetupId, userUid, userName, "pending");

    // 2. Add to meetup's pendingUids
    const meetup = await getFirestoreMeetup(meetupId);
    if (!meetup.pendingUids.includes(userUid)) {
      meetup.pendingUids.push(userUid);
    }
    if (fcmToken) {
      meetup.userFcmTokens[userUid] = fcmToken;
    }

    await updateFirestoreMeetup(meetup, ["pendingUids", "userFcmTokens"]);

    // 3. Send Notification to Host
    if (meetup.hostFcmToken) {
      const bodyText = `${userName} muốn xin vào kèo "${meetup.title}" chơi game ${meetup.game} của bạn.`;
      sendFCMNotification(
        meetup.hostFcmToken,
        "🎯 Yêu cầu tham gia kèo mới!",
        bodyText,
        `/#/manage/${meetupId}`
      ).catch((e) => console.error("FCM notify host failed:", e));
    }

    return c.json({
      success: true,
      message: "Đã gửi yêu cầu tham gia kèo và thông báo tới Host!",
    });
  } catch (err: any) {
    c.status(500);
    return c.json({ error: err.message });
  }
});

// Approve Member API
app.post("/api/meetups/approve", async (c) => {
  let body: any;
  try {
    body = await c.req.json();
  } catch {
    c.status(400);
    return c.json({ error: "Invalid JSON body" });
  }

  const { meetupId, playerUid } = body;
  if (!meetupId || !playerUid) {
    c.status(400);
    return c.json({ error: "Missing required fields" });
  }

  try {
    // 1. Update request status to approved
    await updateFirestoreRequestStatus(meetupId, playerUid, "approved");

    // 2. Transfer from pendingUids to approvedPendingUids
    const meetup = await getFirestoreMeetup(meetupId);
    meetup.pendingUids = meetup.pendingUids.filter((uid) => uid !== playerUid);
    if (!meetup.approvedPendingUids.includes(playerUid)) {
      meetup.approvedPendingUids.push(playerUid);
    }

    await updateFirestoreMeetup(meetup, ["approvedPendingUids", "pendingUids"]);

    // 3. Send Push Notification to Player
    const playerToken = meetup.userFcmTokens[playerUid];
    if (playerToken) {
      const host = meetup.hostName || "Host";
      const bodyText = `Bạn đã được duyệt tham gia kèo "${meetup.title}" chơi game ${meetup.game} của ${host}! Hãy xác nhận tham gia kèo chính thức.`;
      sendFCMNotification(
        playerToken,
        "🎉 Yêu cầu đã được duyệt!",
        bodyText,
        `/#/manage/${meetupId}`
      ).catch((e) => console.error("FCM notify player failed:", e));
    }

    return c.json({
      success: true,
      message: "Đã duyệt thành viên và gửi thông báo!",
    });
  } catch (err: any) {
    c.status(500);
    return c.json({ error: err.message });
  }
});

// Confirm Participation API
app.post("/api/meetups/confirm", async (c) => {
  let body: any;
  try {
    body = await c.req.json();
  } catch {
    c.status(400);
    return c.json({ error: "Invalid JSON body" });
  }

  const { meetupId, userUid, userName } = body;
  if (!meetupId || !userUid || !userName) {
    c.status(400);
    return c.json({ error: "Missing required fields" });
  }

  try {
    // 1. Fetch meetup
    const meetup = await getFirestoreMeetup(meetupId);

    // 2. Transition from approvedPendingUids to approvedUids and increment player count
    meetup.approvedPendingUids = meetup.approvedPendingUids.filter((uid) => uid !== userUid);
    if (!meetup.approvedUids.includes(userUid)) {
      meetup.approvedUids.push(userUid);
      meetup.playersCount++;
    }

    await updateFirestoreMeetup(meetup, ["approvedUids", "approvedPendingUids", "playersCount"]);

    // 3. Notify everyone else (approved members + host)
    const targets = new Set<string>();
    for (const uid of meetup.approvedUids) {
      if (uid !== userUid) targets.add(uid);
    }
    if (meetup.hostUID && meetup.hostUID !== userUid) {
      targets.add(meetup.hostUID);
    }

    const bodyText = `${userName} đã xác nhận tham gia kèo "${meetup.title}" chơi game ${meetup.game}.`;
    for (const uid of targets) {
      let token = meetup.userFcmTokens[uid];
      if (!token && uid === meetup.hostUID) {
        token = meetup.hostFcmToken;
      }
      if (token) {
        sendFCMNotification(
          token,
          "➕ Kèo có thêm người chơi mới!",
          bodyText,
          `/#/manage/${meetupId}`
        ).catch((e) => console.error("FCM notify confirmed user failed:", e));
      }
    }

    return c.json({
      success: true,
      message: "Đã xác nhận vào kèo chính thức và thông báo tới mọi người!",
    });
  } catch (err: any) {
    c.status(500);
    return c.json({ error: err.message });
  }
});

// Leave / Kick Member API
app.post("/api/meetups/leave", async (c) => {
  let body: any;
  try {
    body = await c.req.json();
  } catch {
    c.status(400);
    return c.json({ error: "Invalid JSON body" });
  }

  const { meetupId, playerUid, playerName, isKick } = body;
  if (!meetupId || !playerUid || !playerName) {
    c.status(400);
    return c.json({ error: "Missing required fields" });
  }

  try {
    // 1. Delete request subcollection record
    await deleteFirestoreRequest(meetupId, playerUid);

    // 2. Load meetup and filter arrays
    const meetup = await getFirestoreMeetup(meetupId);

    const wasApproved = meetup.approvedUids.includes(playerUid);
    meetup.approvedUids = meetup.approvedUids.filter((uid) => uid !== playerUid);
    meetup.pendingUids = meetup.pendingUids.filter((uid) => uid !== playerUid);
    meetup.approvedPendingUids = meetup.approvedPendingUids.filter((uid) => uid !== playerUid);

    if (wasApproved && meetup.playersCount > 1) {
      meetup.playersCount--;
    }

    await updateFirestoreMeetup(meetup, [
      "approvedUids",
      "approvedPendingUids",
      "pendingUids",
      "playersCount",
    ]);

    // 3. Send notifications
    if (isKick) {
      // Host kick player -> notify player
      const playerToken = meetup.userFcmTokens[playerUid];
      if (playerToken) {
        const bodyText = `Host đã xóa bạn khỏi danh sách tham gia kèo "${meetup.title}".`;
        sendFCMNotification(playerToken, "✕ Bạn đã bị xóa khỏi kèo", bodyText, "/").catch((e) =>
          console.error("FCM notify kicked player failed:", e)
        );
      }

      // Notify other members
      const targets = new Set<string>();
      for (const uid of meetup.approvedUids) {
        if (uid !== playerUid) targets.add(uid);
      }
      if (meetup.hostUID && meetup.hostUID !== playerUid) {
        targets.add(meetup.hostUID);
      }

      const bodyText = `${playerName} đã không còn tham gia kèo "${meetup.title}".`;
      for (const uid of targets) {
        let token = meetup.userFcmTokens[uid];
        if (!token && uid === meetup.hostUID) {
          token = meetup.hostFcmToken;
        }
        if (token) {
          sendFCMNotification(token, "👋 Thành viên đã rời kèo", bodyText, `/#/manage/${meetupId}`).catch(
            (e) => console.error("FCM notify member left failed:", e)
          );
        }
      }
    } else {
      // Player left voluntarily -> notify host and other members
      const targets = new Set<string>();
      for (const uid of meetup.approvedUids) {
        if (uid !== playerUid) targets.add(uid);
      }
      if (meetup.hostUID && meetup.hostUID !== playerUid) {
        targets.add(meetup.hostUID);
      }

      const bodyText = `${playerName} đã rời khỏi kèo "${meetup.title}".`;
      for (const uid of targets) {
        let token = meetup.userFcmTokens[uid];
        if (!token && uid === meetup.hostUID) {
          token = meetup.hostFcmToken;
        }
        if (token) {
          sendFCMNotification(token, "🚪 Thành viên rời kèo", bodyText, `/#/manage/${meetupId}`).catch(
            (e) => console.error("FCM notify member left failed:", e)
          );
        }
      }
    }

    return c.json({
      success: true,
      message: "Đã xử lý rời kèo/kick và gửi thông báo thành công!",
    });
  } catch (err: any) {
    c.status(500);
    return c.json({ error: err.message });
  }
});

// Start the Node HTTP server on port 8080
const port = 8080;
serve({
  fetch: app.fetch,
  port: port,
});

console.log(`Server Hono/Node.js đang chạy tại http://localhost:${port}...`);
export default app;
