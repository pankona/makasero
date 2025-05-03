# Makasero Web Frontend

This is the frontend for the Makasero project, a hybrid Next.js and Vite application.

## Setup

To set up the project, run:

```bash
npm run setup
```

This will install all dependencies with the `--legacy-peer-deps` flag to handle React version conflicts.

## Development

To start the development server:

```bash
npm run dev
```

This will start the Next.js development server.

## Building

To build the frontend for production:

```bash
npm run build
```

This will:
1. Create the output directory at `~/.makasero/web-frontend` if it doesn't exist
2. Build the Vite application with optimized static files
3. Output the static files to `~/.makasero/web-frontend`

The backend will automatically serve these static files from `~/.makasero/web-frontend` by default, or you can specify a custom directory using the `-static-dir` flag when running the backend.

## Project Structure

- `src/` - Contains the Vite/React application code
- `app/` - Contains the Next.js application code
- `public/` - Static assets
- `build.js` - Custom build script for generating static files
- `vite.config.js` - Vite configuration for building the frontend
