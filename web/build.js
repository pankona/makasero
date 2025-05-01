const { execSync } = require("child_process")
const fs = require("fs")
const path = require("path")

// Ensure the public directory exists
const publicDir = path.resolve(__dirname, "public")
if (!fs.existsSync(publicDir)) {
  fs.mkdirSync(publicDir)
}

// Run the Vite build
console.log("Building Vite application...")
execSync("npx tsc && npx vite build", { stdio: "inherit" })

// Run the Next.js build
console.log("Building Next.js application...")
execSync("npx next build", { stdio: "inherit" })

console.log("Build completed successfully!")
