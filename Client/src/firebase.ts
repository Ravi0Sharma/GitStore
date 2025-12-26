import { initializeApp } from 'firebase/app';
import { getAuth, createUserWithEmailAndPassword, signInWithEmailAndPassword, signOut, setPersistence, browserLocalPersistence, User, AuthError } from 'firebase/auth';

const firebaseConfig = {
  apiKey: import.meta.env.VITE_FIREBASE_APIKEY,
  authDomain: import.meta.env.VITE_FIREBASE_AUTH_DOMAIN,
  projectId: import.meta.env.VITE_FIREBASE_PROJECT_ID,
  storageBucket: import.meta.env.VITE_FIREBASE_STORAGE_BUCKET,
  messagingSenderId: import.meta.env.VITE_FIREBASE_MESSAGE_SENDER_ID,
  appId: import.meta.env.VITE_FIREBASE_APP_ID,
  measurementId: import.meta.env.VITE_FIREBASE_MEASUREMENT_ID,
};

// Validate Firebase config
const requiredEnvVars = ['VITE_FIREBASE_APIKEY', 'VITE_FIREBASE_AUTH_DOMAIN', 'VITE_FIREBASE_PROJECT_ID', 'VITE_FIREBASE_APP_ID'];
const missingVars = requiredEnvVars.filter(varName => !import.meta.env[varName]);

if (missingVars.length > 0) {
  console.error('❌ Firebase config saknar:', missingVars);
  console.error('Kontrollera att .env-filen finns och innehåller alla VITE_FIREBASE_* variabler');
}

const app = initializeApp(firebaseConfig);
export const firebaseAuth = getAuth(app);

// Set persistence to local storage (async)
setPersistence(firebaseAuth, browserLocalPersistence).catch((error) => {
  console.error('Firebase setPersistence error:', error);
});

export async function signUp(email: string, password: string) {
  try {
    const userCredential = await createUserWithEmailAndPassword(firebaseAuth, email, password);
    return { user: userCredential.user, error: null };
  } catch (error: unknown) {
    // Log full error details for debugging
    console.error('Firebase signUp error:', {
      code: (error as AuthError)?.code,
      message: (error as AuthError)?.message,
      fullError: error,
    });
    return { user: null, error: (error as AuthError)?.code || 'unknown' };
  }
}

export async function signIn(email: string, password: string) {
  try {
    const userCredential = await signInWithEmailAndPassword(firebaseAuth, email, password);
    return { user: userCredential.user, error: null };
  } catch (error: unknown) {
    // Log full error details for debugging
    console.error('Firebase signIn error:', {
      code: (error as AuthError)?.code,
      message: (error as AuthError)?.message,
      fullError: error,
    });
    return { user: null, error: (error as AuthError)?.code || 'unknown' };
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

export function toAuthErrorMessage(error: unknown): string {
  // Handle Firebase AuthError
  if (error && typeof error === 'object' && 'code' in error) {
    const authError = error as AuthError;
    const errorCode = authError.code;
    
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
      case 'auth/invalid-credential':
        return 'Ogiltiga inloggningsuppgifter. Kontrollera email och lösenord.';
      case 'auth/operation-not-allowed':
        return 'Denna inloggningsmetod är inte aktiverad. Kontrollera Firebase Console.';
      case 'auth/missing-password':
        return 'Lösenord saknas.';
      case 'auth/invalid-api-key':
        return 'Ogiltig Firebase API-nyckel. Kontrollera .env-filen.';
      case 'auth/app-not-authorized':
        return 'Firebase-appen är inte auktoriserad. Kontrollera Firebase Console.';
      default:
        // Show Firebase's actual error message if available, otherwise show code
        return authError.message || `Firebase-fel: ${errorCode}`;
    }
  }
  
  // Handle string error codes (backward compatibility)
  if (typeof error === 'string') {
    return toAuthErrorMessage({ code: error } as AuthError);
  }
  
  // Fallback for unknown errors
  if (error && typeof error === 'object' && 'message' in error) {
    return (error as { message: string }).message || 'Ett oväntat fel uppstod.';
  }
  
  return 'Ett oväntat fel uppstod. Kontrollera konsolen för mer information.';
}

// Backward compatibility wrapper
export function getFirebaseErrorMessage(errorCode: string): string {
  return toAuthErrorMessage({ code: errorCode } as AuthError);
}

