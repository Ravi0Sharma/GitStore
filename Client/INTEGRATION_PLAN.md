# Plan: Integrera Landing Page i Huvudapp

## Nuvarande Situation

### Struktur
- **Huvudapp**: `/Client/` med minimal setup (Vite + React + TS + Tailwind v4)
- **Landing page**: `/Client/landingpagee/` med full shadcn/radix setup (Tailwind v3)
- **Problem**: Två separata projekt, landing page använder `@/` alias som inte är konfigurerat i huvudapp

### Vad Landing Page Faktiskt Använder

**Komponenter som används:**
- ✅ `Navbar`, `HeroSection`, `TimelineSection`, `IssuesSection`, `IssueManagementSection`, `ViewHistorySection`, `FeatureCard`, `Starfield` - **Inga shadcn UI-komponenter direkt**
- ✅ `lucide-react` - Icons (används överallt)
- ✅ `clsx` + `tailwind-merge` - För `cn()` utility (används i NavLink och UI-komponenter)
- ✅ `@tanstack/react-query` - Används i App.tsx (QueryClientProvider)
- ✅ `react-router-dom` - Används i App.tsx (men kan tas bort om vi bara visar landing)
- ⚠️ `@/components/ui/toaster`, `sonner`, `tooltip` - Används i App.tsx men kanske inte nödvändigt för landing

**UI-komponenter som INTE används direkt av landing:**
- Alla shadcn UI-komponenter i `ui/` mappen används bara internt av varandra, inte av landing-komponenterna

## Planerad Lösning

### A) Stabilisera Entry + Integration

#### 1. Säkerställ Entry Point
- ✅ `src/main.tsx` är redan korrekt - renderar bara `<App />`
- ✅ `src/App.tsx` behöver uppdateras för att rendera landing

#### 2. Skapa Landing Modulstruktur
```
src/
  landing/
    LandingPage.tsx          # Huvudkomponent som exporterar hela landing-sidan
    components/              # Flytta hit från landingpagee/src/components/
      Navbar.tsx
      HeroSection.tsx
      TimelineSection.tsx
      IssuesSection.tsx
      IssueManagementSection.tsx
      ViewHistorySection.tsx
      FeatureCard.tsx
      Starfield.tsx
      NavLink.tsx
    hooks/                   # Flytta hit från landingpagee/src/hooks/
      use-mobile.tsx
      use-toast.ts
    lib/
      utils.ts               # Flytta hit från landingpagee/src/lib/utils.ts
```

#### 3. Fixa Imports
**Strategi: Konfigurera `@/` alias i Vite + TypeScript**
- Lägg till path alias i `vite.config.js`: `@` → `./src`
- Skapa `tsconfig.json` med samma alias
- Alternativ: Byt alla `@/` till relativa imports (mer arbete, mindre risk)

**Val: Konfigurera alias** (minst förändringar)

### B) Tailwind Sanity

#### 4. Tailwind Konfiguration
- **Nuvarande**: Huvudapp använder Tailwind v4 via `@tailwindcss/vite` plugin
- **Landing page**: Använder Tailwind v3 med `tailwind.config.js`
- **Lösning**: 
  - Migrera landing-styling till Tailwind v4 format
  - Eller: Använd Tailwind v3 i huvudapp (enklare för nu)
  - **Val: Använd Tailwind v3** (landing är redan byggd för det)

**Tailwind Config behöver:**
```js
content: [
  "./index.html",
  "./src/**/*.{ts,tsx}",
  "./src/landing/**/*.{ts,tsx}"
]
```

**CSS:**
- Flytta `landingpagee/src/index.css` innehåll till `src/index.css`
- Säkerställ att Tailwind directives finns: `@tailwind base/components/utilities`

### C) Dependencies Cleanup

#### 5. Minimal package.json

**Behåll:**
- `react`, `react-dom` - Core
- `typescript`, `@types/react`, `@types/react-dom` - TypeScript
- `vite`, `@vitejs/plugin-react` - Build tool
- `tailwindcss`, `autoprefixer`, `postcss` - Tailwind v3
- `lucide-react` - Icons (används överallt)
- `clsx`, `tailwind-merge` - Utility functions
- `@tanstack/react-query` - Används i App.tsx (kan tas bort om vi förenklar App.tsx)

**Ta bort (inte används av landing):**
- Alla `@radix-ui/*` paket (används bara av UI-komponenter som inte används)
- `react-router-dom` (om vi tar bort routing)
- `react-hook-form`, `@hookform/resolvers`, `zod` (inte används)
- `date-fns`, `react-day-picker` (inte används)
- `embla-carousel-react`, `cmdk`, `input-otp` (inte används)
- `recharts` (inte används)
- `sonner` (används i App.tsx men kanske inte nödvändigt)
- `next-themes` (inte används)
- `react-resizable-panels`, `vaul` (inte används)
- `class-variance-authority` (används bara av UI-komponenter)
- `tailwindcss-animate` (kan behövas för animations)

**UI-komponenter:**
- Behåll endast UI-komponenter som faktiskt används:
  - `toaster.tsx`, `toast.tsx` (om vi behåller toast)
  - `tooltip.tsx` (om vi behåller TooltipProvider)
  - `sonner.tsx` (om vi behåller Sonner)
- Ta bort alla andra UI-komponenter

### D) Förenkla App.tsx

#### 6. Minimal App.tsx
**Alternativ 1: Ingen routing (enklast)**
```tsx
import LandingPage from './landing/LandingPage';
import './index.css';

function App() {
  return <LandingPage />;
}
```

**Alternativ 2: Behåll providers (om toast/tooltip behövs)**
```tsx
import LandingPage from './landing/LandingPage';
import { TooltipProvider } from './landing/components/ui/tooltip';
import './index.css';

function App() {
  return (
    <TooltipProvider>
      <LandingPage />
    </TooltipProvider>
  );
}
```

**Val: Alternativ 1** (enklast, ta bort alla providers om de inte används)

## Steg-för-Steg Implementation

### Steg 1: Konfigurera Path Alias
1. Uppdatera `vite.config.js` med path resolve
2. Skapa `tsconfig.json` med paths

### Steg 2: Flytta Landing Komponenter
1. Skapa `src/landing/` struktur
2. Flytta komponenter från `landingpagee/src/components/` → `src/landing/components/`
3. Flytta hooks → `src/landing/hooks/`
4. Flytta lib/utils.ts → `src/landing/lib/utils.ts`

### Steg 3: Skapa LandingPage.tsx
1. Skapa `src/landing/LandingPage.tsx` baserat på `landingpagee/src/pages/Index.tsx`
2. Uppdatera imports till nya paths

### Steg 4: Uppdatera App.tsx
1. Importera och rendera `<LandingPage />`
2. Ta bort routing och providers om inte nödvändigt

### Steg 5: Tailwind Setup
1. Installera Tailwind v3 dependencies
2. Skapa `tailwind.config.js` med korrekt content paths
3. Skapa `postcss.config.js`
4. Uppdatera `src/index.css` med Tailwind directives + landing styles

### Steg 6: Cleanup Dependencies
1. Analysera vilka UI-komponenter som faktiskt behövs
2. Ta bort onödiga dependencies från package.json
3. Ta bort onödiga UI-komponenter från `src/landing/components/ui/`

### Steg 7: Testa
1. `npm install`
2. `npm run dev` - ska starta utan errors
3. `npm run build` - ska bygga utan errors
4. Verifiera att landing-sidan renderas korrekt

## Risker & Överväganden

1. **Tailwind v3 vs v4**: Landing är byggd för v3, huvudapp använder v4. Lösning: Använd v3.
2. **Path alias**: Om alias inte fungerar, kan vi byta till relativa imports.
3. **UI-komponenter**: Många UI-komponenter används inte direkt men kan behövas indirekt. Lösning: Ta bort stegvis och testa.
4. **CSS Variables**: Landing använder CSS custom properties som måste finnas i index.css.

## Definition of Done Checklist

- [ ] `npm install` fungerar utan errors
- [ ] `npm run dev` startar och visar landing-sidan
- [ ] `npm run build` fungerar utan errors
- [ ] Landing-sidan renderas korrekt med alla sektioner
- [ ] Alla styles (gradients, animations, etc.) fungerar
- [ ] package.json innehåller bara nödvändiga dependencies
- [ ] Inga trasiga imports
- [ ] Path alias `@/` fungerar korrekt

---

**Godkänn denna plan innan implementation börjar?**
