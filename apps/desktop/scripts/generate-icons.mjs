import { readFile, writeFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import sharp from "sharp";
import pngToIco from "png-to-ico";

const root = path.join(path.dirname(fileURLToPath(import.meta.url)), "..");
const iconsDir = path.join(root, "icons");
const svgPath = path.join(iconsDir, "icon.svg");
const pngPath = path.join(iconsDir, "icon.png");
const icoPath = path.join(iconsDir, "icon.ico");

const svg = await readFile(svgPath);

await sharp(svg, { density: 384 })
  .resize(1024, 1024)
  .png()
  .toFile(pngPath);

const sizes = [256, 128, 64, 48, 32, 16];
const pngBuffers = await Promise.all(
  sizes.map((size) => sharp(pngPath).resize(size, size).png().toBuffer()),
);

await writeFile(icoPath, await pngToIco(pngBuffers));

console.log("Generated", path.relative(root, pngPath), "and", path.relative(root, icoPath));
