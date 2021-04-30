const supportImages = []
const pendingTests = []
const testImages = {
  /**
     * Test JPEG image, encoded as base64.
     */
  'image/jpeg':
        '/9j/4AAQSkZJRgABAQEASABIAAD/2wBDAAMCAgMCAgMDAwMEAwMEBQgFBQQEBQoH' +
        'BwYIDAoMDAsKCwsNDhIQDQ4RDgsLEBYQERMUFRUVDA8XGBYUGBIUFRT/2wBDAQME' +
        'BAUEBQkFBQkUDQsNFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQU' +
        'FBQUFBQUFBQUFBQUFBT/wAARCAABAAEDAREAAhEBAxEB/8QAFAABAAAAAAAAAAA' +
        'AAAAAAAAACf/EABQQAQAAAAAAAAAAAAAAAAAAAAD/xAAUAQEAAAAAAAAAAAAAAA' +
        'AAAAAA/8QAFBEBAAAAAAAAAAAAAAAAAAAAAP/aAAwDAQACEQMRAD8AVMH/2Q==',

  /**
     * Test PNG image, encoded as base64.
     */
  'image/png':
        'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABAQMAAAAl21bKAAAAA1BMVEX///+nxBvI' +
        'AAAACklEQVQI12NgAAAAAgAB4iG8MwAAAABJRU5ErkJggg==',

  /**
     * Test WebP image, encoded as base64.
     */
  'image/webp': 'UklGRhoAAABXRUJQVlA4TA0AAAAvAAAAEAcQERGIiP4HAA=='

}

for (const key in testImages) {
  const imageTest = new Promise(function(resolve, reject) {
    const item = testImages[key]
    const image = new Image()
    image.src = 'data:' + key + ';base64,' + item
    image.onload = image.onerror = function imageTestComplete() {
      // Image format is supported if successfully decoded
      if (image.width === 1 && image.height === 1) { supportImages.push(key) }
      // Test is complete
      resolve()
    }
  })
  pendingTests.push(imageTest)
}

export function getSupportedMimetypes() {
  return new Promise(function(resolve, reject) {
    Promise.all(pendingTests).then(() => {
      resolve(supportImages)
    })
  })
}
