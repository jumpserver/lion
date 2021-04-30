import Guacamole from 'guacamole-common-js'

export function getSupportedGuacAudios() {
  return Guacamole.AudioPlayer.getSupportedTypes()
}

