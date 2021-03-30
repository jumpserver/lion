export function sanitizeFilename(filename) {
  return filename.replace(/[\\\/]+/g, '_')
}