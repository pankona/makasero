# Issues for Improving devcontainer.json

The following issues have been created to break down the work for improving the devcontainer.json setup:

- [#38](https://github.com/pankona/makasero/issues/38) - Modify makasero-web-backend to serve static frontend files
- [#39](https://github.com/pankona/makasero/issues/39) - Create build process for frontend and generate static files
- [#40](https://github.com/pankona/makasero/issues/40) - Create installation process for makasero-web-backend binary
- [#41](https://github.com/pankona/makasero/issues/41) - Update devcontainer.json to use installed binary and pre-built frontend

These issues outline the steps needed to improve the development environment by:
1. Using a pre-installed binary instead of `go run`
2. Serving pre-built frontend files from the backend instead of using `npm run dev`
