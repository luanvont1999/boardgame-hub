import { 
  doc, 
  collection, 
  setDoc, 
  deleteDoc, 
  updateDoc, 
  arrayUnion, 
  arrayRemove, 
  increment, 
  serverTimestamp,
  onSnapshot
} from 'firebase/firestore';
import type { User } from 'firebase/auth';
import { db } from './firebase';

export interface MeetupRequest {
  uid: string;
  name: string;
  status: 'pending' | 'approved' | 'rejected';
  createdAt?: any;
}

/**
 * Checks if a user is an approved member or host of a meetup
 */
export function isApprovedMember(meetup: any, userUid: string | null | undefined): boolean {
  if (!userUid || !meetup) return false;
  const hostUid = meetup.hostUid || meetup.host_uid;
  if (userUid === hostUid) return true;
  if (Array.isArray(meetup.approvedUids) && meetup.approvedUids.includes(userUid)) {
    return true;
  }
  return false;
}

/**
 * Checks if user is the creator/host of the meetup
 */
export function isHost(meetup: any, userUid: string | null | undefined): boolean {
  if (!userUid || !meetup) return false;
  const hostUid = meetup.hostUid || meetup.host_uid;
  return userUid === hostUid;
}

/**
 * Player sends a join request to a meetup
 */
export async function requestToJoin(meetupId: string, user: User) {
  const reqRef = doc(db, 'meetups', meetupId, 'requests', user.uid);
  const meetupRef = doc(db, 'meetups', meetupId);
  const name = user.displayName || user.email || 'Thành viên';
  
  await setDoc(reqRef, {
    uid: user.uid,
    name,
    status: 'pending',
    createdAt: serverTimestamp()
  });

  await updateDoc(meetupRef, {
    pendingUids: arrayUnion(user.uid)
  });
}

/**
 * Player cancels their pending join request
 */
export async function cancelJoinRequest(meetupId: string, userUid: string) {
  const reqRef = doc(db, 'meetups', meetupId, 'requests', userUid);
  const meetupRef = doc(db, 'meetups', meetupId);

  await deleteDoc(reqRef);
  await updateDoc(meetupRef, {
    pendingUids: arrayRemove(userUid)
  });
}

/**
 * Host approves a pending player to join the meetup
 */
export async function approveMember(meetupId: string, playerUid: string) {
  const reqRef = doc(db, 'meetups', meetupId, 'requests', playerUid);
  const meetupRef = doc(db, 'meetups', meetupId);

  await updateDoc(reqRef, { status: 'approved' });
  await updateDoc(meetupRef, {
    approvedUids: arrayUnion(playerUid),
    pendingUids: arrayRemove(playerUid),
    playersCount: increment(1)
  });
}

/**
 * Host rejects a pending player request
 */
export async function rejectMember(meetupId: string, playerUid: string) {
  const reqRef = doc(db, 'meetups', meetupId, 'requests', playerUid);
  const meetupRef = doc(db, 'meetups', meetupId);

  await updateDoc(reqRef, { status: 'rejected' });
  await updateDoc(meetupRef, {
    pendingUids: arrayRemove(playerUid)
  });
}

/**
 * Host kicks an approved member OR member leaves the meetup
 */
export async function kickOrLeaveMember(meetupId: string, playerUid: string) {
  const reqRef = doc(db, 'meetups', meetupId, 'requests', playerUid);
  const meetupRef = doc(db, 'meetups', meetupId);

  await deleteDoc(reqRef);
  await updateDoc(meetupRef, {
    approvedUids: arrayRemove(playerUid),
    pendingUids: arrayRemove(playerUid), // defensive cleanup
    playersCount: increment(-1)
  });
}

/**
 * Realtime subscription to all requests of a meetup
 */
export function subscribeToMeetupRequests(
  meetupId: string, 
  callback: (requests: MeetupRequest[]) => void
) {
  const reqsRef = collection(db, 'meetups', meetupId, 'requests');
  return onSnapshot(reqsRef, (snapshot) => {
    const list = snapshot.docs.map(docSnap => ({
      uid: docSnap.id,
      ...docSnap.data()
    })) as MeetupRequest[];
    callback(list);
  }, (err) => {
    console.error('[MeetupService] Requests listener error:', err);
  });
}
