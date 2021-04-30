import Guacamole from 'guacamole-common-js'

export function getSupportedGuacVideos() {
  return Guacamole.VideoPlayer.getSupportedTypes()
}
