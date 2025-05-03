const { execSync } = require("child_process")
const fs = require("fs")
const path = require("path")
const os = require("os")

// Define output directory
const outputDir = path.join(os.homedir(), ".makasero", "web-frontend")

// Ensure the output directory exists
if (!fs.existsSync(outputDir)) {
  console.log(`Creating output directory: ${outputDir}`)
  fs.mkdirSync(outputDir, { recursive: true })
}

// Ensure the public directory exists
const publicDir = path.resolve(__dirname, "public")
if (!fs.existsSync(publicDir)) {
  fs.mkdirSync(publicDir)
}

// Run the Vite build
console.log("Building Vite application...")
try {
  execSync("npx vite build", { stdio: "inherit" })
  console.log("Vite build completed successfully!")
} catch (error) {
  console.error("Vite build failed:", error.message)
  process.exit(1)
}

console.log("Build completed successfully!")
console.log(`Static files have been generated in: ${outputDir}`)
