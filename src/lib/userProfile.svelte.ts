import { doc, getDoc, setDoc } from "firebase/firestore";
import { db, auth } from "./firebase";
import { onAuthStateChanged } from "firebase/auth";

export interface UserProfile {
  uid: string;
  displayName: string;
  bio: string;
  favoriteCategories: string[];
  loaded: boolean;
}

class UserProfileState {
  profile = $state<UserProfile>({
    uid: "",
    displayName: "",
    bio: "",
    favoriteCategories: [],
    loaded: false,
  });

  isLoading = $state(false);

  constructor() {
    if (typeof window !== "undefined") {
      onAuthStateChanged(auth, async (user) => {
        if (user) {
          const isUserChanged = this.profile.uid !== user.uid;
          this.profile.uid = user.uid;
          // Fetch if user changed or not loaded yet
          if (isUserChanged || !this.profile.loaded) {
            await this.fetchProfile(user.uid, user.displayName);
          }
        } else {
          // Reset state on sign out
          this.profile = {
            uid: "",
            displayName: "",
            bio: "",
            favoriteCategories: [],
            loaded: false,
          };
        }
      });
    }
  }

  async fetchProfile(uid: string, defaultDisplayName?: string | null) {
    this.isLoading = true;
    try {
      const docRef = doc(db, "users", uid);
      const docSnap = await getDoc(docRef);
      if (docSnap.exists()) {
        const data = docSnap.data();
        this.profile.displayName = data.displayName || defaultDisplayName || "";
        this.profile.bio = data.bio || "";
        this.profile.favoriteCategories = data.favoriteCategories || [];
      } else {
        this.profile.displayName = defaultDisplayName || "";
        this.profile.bio = "";
        this.profile.favoriteCategories = [];
      }
      this.profile.loaded = true;
    } catch (err) {
      console.error("Lỗi tải thông tin user profile:", err);
    } finally {
      this.isLoading = false;
    }
  }

  async updateProfile(data: { displayName: string; bio: string; favoriteCategories: string[] }) {
    if (!this.profile.uid) return;
    this.isLoading = true;
    try {
      const docRef = doc(db, "users", this.profile.uid);
      await setDoc(docRef, {
        displayName: data.displayName,
        bio: data.bio,
        favoriteCategories: data.favoriteCategories,
        updatedAt: new Date(),
      }, { merge: true });

      // Update local state
      this.profile.displayName = data.displayName;
      this.profile.bio = data.bio;
      this.profile.favoriteCategories = data.favoriteCategories;
      this.profile.loaded = true;
    } catch (err) {
      console.error("Lỗi cập nhật user profile:", err);
      throw err;
    } finally {
      this.isLoading = false;
    }
  }
}

export const userProfileState = new UserProfileState();
