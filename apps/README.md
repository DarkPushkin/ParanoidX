# The Isle — Client Applications

## Structure

```
apps/
├── royal_app/      → Royal Node Control (admin panel for King family)
│   └── Flutter app (macOS + Linux Desktop, iOS + Android)
├── isle_app/       → The Isle (citizen app for all island services)
│   └── Flutter app (iOS + Android, Web + Desktop)
└── shared/
    ├── api_client/ → Typed HTTP client for simplex-node REST API
    ├── models/     → Shared Dart data models
    └── widgets/    → Reusable Flutter UI components
```

## Prerequisites

- Flutter SDK 3.29+ (`flutter --version`)
- Go backend running on localhost:8080 (or configured base URL)

## Getting Started

```bash
# Build shared packages
cd apps/shared/api_client && dart pub get
cd apps/shared/models && dart pub get
cd apps/shared/widgets && flutter pub get

# Run Royal App
cd apps/royal_app && flutter run -d linux

# Run Isle App
cd apps/isle_app && flutter run -d android
```

## Architecture

Both apps communicate with the simplex-node backend via the shared `api_client` package.
All data models are defined in `shared/models` and shared between apps.
Reusable UI components live in `shared/widgets`.

See THEPLAN.md (Track 5) and docs/EVOLUTION-PLAN.md (Section 6) for full timeline.
