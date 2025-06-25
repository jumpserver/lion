//@ts-ignore 
import Guacamole from 'guacamole-common-js'

const supportImages:any[] = []
const pendingTests:any[] = []
const testImages:any = {
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

// Use Object.entries for better iteration over key-value pairs
Object.entries(testImages).forEach(([mimeType, base64Data]) => {
  const imageTest = new Promise<void>((resolve) => {
    const image = new Image();
    
    // Set up handlers before setting src to avoid race conditions
    image.onload = () => {
      // Image format is supported if successfully decoded with correct dimensions
      if (image.width === 1 && image.height === 1) {
        supportImages.push(mimeType);
      }
      resolve();
    };
    
    // Handle errors separately for better debugging
    image.onerror = () => {
      console.debug(`Format ${mimeType} not supported`);
      resolve(); // Still resolve to continue testing other formats
    };
    
    // Set source to trigger loading
    image.src = `data:${mimeType};base64,${base64Data}`;
  });
  
  pendingTests.push(imageTest);
});

export async function getSupportedMimetypes() {
  return Promise.all(pendingTests).then(() => supportImages);
}


export function getSupportedGuacVideos() {
  return Guacamole.VideoPlayer.getSupportedTypes()
}

export function getSupportedGuacAudios() {
  return Guacamole.AudioPlayer.getSupportedTypes()
}

