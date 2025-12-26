import { initializeApp } from 'firebase/app';
import { getAuth, createUserWithEmailAndPassword, signInWithEmailAndPassword, signOut, setPersistence, browserLocalPersistence, User } from 'firebase/auth';

const firebaseConfig = {
  apiKey: import.meta.env.VITE_FIREBASE_APIKEY,
  authDomain: import.meta.env.VITE_FIREBASE_AUTH_DOMAIN,
  projectId: import.meta.env.VITE_FIREBASE_PROJECT_ID,
  storageBucket: import.meta.env.VITE_FIREBASE_STORAGE_BUCKET,
  messagingSenderId: import.meta.env.VITE_FIREBASE_MESSAGE_SENDER_ID,
  appId: import.meta.env.VITE_FIREBASE_APP_ID,
  measurementId: import.meta.env.VITE_FIREBASE_MEASUREMENT_ID,
};

const app = initializeApp(firebaseConfig);
export const firebaseAuth = getAuth(app);

// Set persistence to local storage
setPersistence(firebaseAuth, browserLocalPersistence);

export async function signUp(email: string, password: string) {
  try {
    const userCredential = await createUserWithEmailAndPassword(firebaseAuth, email, password);
    return { user: userCredential.user, error: null };
  } catch (error: any) {
    return { user: null, error: error.code };
  }
}

export async function signIn(email: string, password: string) {
  try {
    const userCredential = await signInWithEmailAndPassword(firebaseAuth, email, password);
    return { user: userCredential.user, error: null };
  } catch (error: any) {
    return { user: null, error: error.code };
  }
}

export async function signOutUser() {
  try {
    await signOut(firebaseAuth);
    return { error: null };
  } catch (error: any) {
    return { error: error.code };
  }
}

export function getFirebaseErrorMessage(errorCode: string): string {
  switch (errorCode) {
    case 'auth/email-already-in-use':
      return 'Detta e-postadress används redan. Vänligen använd en annan.';
    case 'auth/wrong-password':
      return 'Felaktigt lösenord. Försök igen.';
    case 'auth/user-not-found':
      return 'Ingen användare hittades med denna e-postadress.';
    case 'auth/invalid-email':
      return 'Ogiltig e-postadress. Kontrollera formatet.';
    case 'auth/too-many-requests':
      return 'För många försök. Vänta en stund och försök igen.';
    case 'auth/weak-password':
      return 'Lösenordet är för svagt. Använd minst 6 tecken.';
    case 'auth/network-request-failed':
      return 'Nätverksfel. Kontrollera din internetanslutning.';
    default:
      return 'Ett fel uppstod. Försök igen.';
  }
}

