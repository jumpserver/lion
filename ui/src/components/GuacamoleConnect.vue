<template>
  <el-container>
    <el-main>
      <el-row
        v-loading="loading"
        :element-loading-text="loadingText"
        element-loading-background="#1f1b1b"
      >
        <div :style="containerStyle">
          <div id="displayOuter">
            <div id="displayMiddle">
              <div id="display" />
            </div>
          </div>
        </div>
      </el-row>
    </el-main>
    <RightPanel ref="panel">
      <Settings :settings="settings" :title="$t('Settings')">
        <el-button type="text" class="item-button el-icon-c-scale-to-original">
          {{ $t('Display') }}
        </el-button>
        <div class="content"> <i class="el-icon-remove-outline" @click="decreaseScale" />
          <span>{{ scaleValue }}%</span>
          <i class="el-icon-circle-plus-outline" @click="increaseScale" />
          <el-form label-position="left">
            <el-form-item :label="$t('AutoFit')">
              <el-switch v-model="autoFit" />
            </el-form-item>
          </el-form>
        </div>
      </Settings>
    </RightPanel>
    <GuacClipboard
      v-if="clipboardInited"
      ref="clipboard"
      :visible.sync="clipboardDrawer"
      :client="client"
      :tunnel="tunnel"
      @closeDrawer="onCloseDrawer"
    />
    <GuacFileSystem
      v-if="fileSystemInited"
      ref="fileSystem"
      :client="client"
      :tunnel="tunnel"
      :show.sync="fileDrawer"
      @closeDrawer="onCloseDrawer"
    />
    <el-dialog
      :title="shareTitle"
      :visible.sync="shareDialogVisible"
      width="30%"
      :close-on-press-escape="false"
      :close-on-click-modal="false"
      :modal="false"
      class="share-dialog"
      center
      @close="shareDialogClosed"
    >
      <div v-if="!shareId">
        <el-form v-loading="shareLoading" label-position="left" :model="shareLinkRequest">
          <el-form-item :label="$t('ExpiredTime')">
            <el-select v-model="shareLinkRequest.expiredTime" :placeholder="$t('SelectAction')">
              <el-option
                v-for="item in expiredOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('ActionPerm')">
            <el-select v-model="shareLinkRequest.actionPerm" :placeholder="$t('ActionPerm')">
              <el-option
                v-for="item in actionsPermOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('ShareUser')">
            <el-select
              v-model="shareLinkRequest.users"
              multiple
              filterable
              remote
              reserve-keyword
              :placeholder="$t('GetShareUser')"
              :remote-method="getSessionUser"
              :loading="userLoading"
            >
              <el-option
                v-for="item in userOptions"
                :key="item.id"
                :label="item.name + '(' + item.username + ')'"
                :value="item.id"
              />
            </el-select>
            <div style="color: #d9d1d1;font-size: 12px">{{ $t('ShareUserHelpText') }}</div>
          </el-form-item>
        </el-form>
        <div>
          <el-button class="share-btn" @click="handleShareURlCreated">{{ $t('CreateLink') }}</el-button>
        </div>
      </div>
      <div v-else>
        <el-result icon="success" class="result" :title="$t('CreateSuccess')" />
        <el-form label-position="top">
          <el-form-item :label="$t('LinkAddr')">
            <el-input readonly :value="shareURL" />
          </el-form-item>
          <el-form-item :label="$t('VerifyCode')">
            <el-input readonly :value="shareCode" />
          </el-form-item>
        </el-form>
        <div class="share">
          <el-button class="copy-btn" @click="copyShareURL">{{ $t('CopyLink') }}</el-button>
        </div>
      </div>
    </el-dialog>
  </el-container>
</template>

<script>
import Guacamole from 'guacamole-common-js'
import { getSupportedMimetypes } from '@/utils/image'
import { getSupportedGuacAudios } from '@/utils/audios'
import { getSupportedGuacVideos } from '@/utils/video'
import { getCurrentConnectParams, getURLParams,
  localStorageGet, BASE_URL, CopyTextToClipboard } from '@/utils/common'
import GuacClipboard from './GuacClipboard'
import GuacFileSystem from './GuacFileSystem'
import RightPanel from './RightPanel'
import Settings from './Settings'
import { default as i18n, getLanguage } from '@/i18n'
import { ErrorStatusCodes, ConvertGuacamoleError } from '@/utils'
import { getUserInfo } from '@/api/user'
import { createShareURL, removeShareUser } from '@/api/session'

const pixelDensity = 1
const sideWidth = 0
export default {
  name: 'GuacamoleConnect',
  components: {
    GuacClipboard,
    GuacFileSystem,
    RightPanel,
    Settings
  },
  data() {
    return {
      apiPrefix: '/api',
      wsPrefix: '/lion/ws/connect/',
      dialogFormVisible: false,
      requireParams: [],
      isMenuCollapse: true,
      hasFileSystem: false,
      clipboardDrawer: false,
      fileDrawer: false,
      loading: true,
      session: null,
      tunnelState: '',
      fileSystemInited: false,
      clipboardInited: false,
      loadingText: i18n.t('Connecting') + ' ...',
      clientState: 'Connecting',
      localCursor: false,
      client: null,
      clientProperties: {
        autoFit: true
      },
      tunnel: null,
      displayWidth: window.innerWidth - sideWidth,
      displayHeight: window.innerHeight,
      connected: false,
      clipboardData: {
        type: 'text/plain',
        data: ''
      },
      sink: null,
      keyboard: null,
      combinationKeys: [
        {
          keys: ['65307'],
          name: 'Esc'
        },
        {
          keys: ['65480'],
          name: 'F11'
        },
        {
          keys: ['65507', '65513', '65535'],
          name: 'Ctrl+Alt+Delete'
        },
        {
          keys: ['65507', '65513', '65288'],
          name: 'Ctrl+Alt+Backspace'
        },
        {
          keys: ['65515', '100'],
          name: 'Windows+D'
        },
        {
          keys: ['65515', '101'],
          name: 'Windows+E'
        },
        {
          keys: ['65515', '114'],
          name: 'Windows+R'
        },
        {
          keys: ['65515', '120'],
          name: 'Windows+X'
        },
        {
          keys: ['65515'],
          name: 'Windows'
        },
        {
          keys: ['65513', '65289'],
          name: 'Alt+Tab'
        }
      ],
      remoteAppCombinationKeys: [
        {
          keys: ['65513', '65289'],
          name: 'Alt+Tab'
        }
      ],
      scale: 0.1,
      timeout: null,
      origin: null,
      lunaId: null,
      display: null,
      autoFit: true,
      shareDialogVisible: false,
      shareId: '',
      shareLinkRequest: {
        expiredTime: 10,
        actionPerm: 'writable',
        users: []
      },
      expiredOptions: [
        { label: this.getMinuteLabel(1), value: 1 },
        { label: this.getMinuteLabel(5), value: 5 },
        { label: this.getMinuteLabel(10), value: 10 },
        { label: this.getMinuteLabel(20), value: 20 },
        { label: this.getMinuteLabel(60), value: 60 }
      ],
      actionsPermOptions: [
        { label: this.$t('Writable'), value: 'writable' },
        { label: this.$t('ReadOnly'), value: 'readonly' }
      ],
      userOptions: [],
      userLoading: false,
      shareCode: null,
      shareLoading: false,
      enableShare: true,
      onlineUsersMap: {}
    }
  },
  computed: {
    containerStyle() {
      return {
        height: this.displayHeight + 'px',
        width: this.displayWidth + 'px'
      }
    },
    scaleValue() {
      return Math.floor(this.scale * 100)
    },
    menuDisable: function() {
      return !(this.clientState === 'Connected') || !(this.tunnelState === 'OPEN')
    },
    isRemoteApp: function() {
      return this.session ? this.session.remote_app : false
    },
    title: function() {
      if (this.isRemoteApp) {
        return this.session.remote_app.name
      }
      return this.session.asset.hostname
    },
    settings() {
      const settings = [
        {
          title: this.$t('Clipboard'),
          icon: 'el-icon-document-copy',
          disabled: () => (this.menuDisable || !this.clipboardInited),
          click: () => (this.toggleClipboard())
        },
        {
          title: this.$t('Files'),
          icon: 'el-icon-folder',
          disabled: () => ((!this.hasFileSystem) || this.menuDisable),
          click: () => (this.toggleFileSystem())
        },
        {
          title: this.$t('Share'),
          icon: 'el-icon-share',
          disabled: () => (this.menuDisable || !this.enableShare),
          click: () => (this.shareDialogVisible = !this.shareDialogVisible)
        }
      ]
      if (!this.isRemoteApp) {
        settings.push({
          title: this.$t('Shortcuts'),
          icon: 'el-icon-position',
          disabled: () => (this.menuDisable),
          content: this.combinationKeys,
          itemClick: (keys) => (this.handleKeys(keys))
        })
      } else {
        settings.push({
          title: this.$t('Shortcuts'),
          icon: 'el-icon-position',
          disabled: () => (this.menuDisable),
          content: this.remoteAppCombinationKeys,
          itemClick: (keys) => (this.handleKeys(keys))
        })
      }
      settings.push({
        title: this.$t('User'),
        icon: 'el-icon-s-custom',
        disabled: () => Object.keys(this.onlineUsersMap).length < 1,
        content: Object.values(this.onlineUsersMap).map(item => {
          item.name = item.user
          item.faIcon = item.writable ? 'fa-solid fa-keyboard' : 'fa-solid fa-eye'
          item.iconTip = item.writable ? this.$t('Writable') : this.$t('ReadOnly')
          return item
        }).sort((a, b) => new Date(a.created) - new Date(b.created)),
        itemClick: () => {
        },
        itemActions: [
          {
            faIcon: 'fa-solid fa-trash-can',
            tipText: this.$t('Remove'),
            style: {
              color: '#f56c6c'
            },
            hidden: (user) => {
              return user.primary
            },
            click: (user) => {
              if (user.primary) {
                return
              }
              this.$confirm(this.$t('RemoveShareUserConfirm'))
                .then(() => {
                  this.$log.debug('Remove user', user)
                  removeShareUser(user).then(response => {
                    this.$log.debug('Remove user response: ', response)
                  })
                    .catch(() => {
                      this.$log.debug('not Remove user', user)
                    })
                })
                .catch(() => {
                  this.$log.debug('not Remove user', user)
                })
            }
          }
        ]
      })
      return settings
    },
    shareTitle() {
      return this.shareId ? this.$t('Share') : this.$t('CreateLink')
    },
    shareURL() {
      return this.shareId ? this.generateShareURL() : this.$t('NoLink')
    }
  },
  watch: {
    autoFit: function(val) {
      if (val) {
        this.updateDisplayScale()
      }
    }
  },
  mounted: function() {
    const result = getCurrentConnectParams()
    this.apiPrefix = result['api']
    this.wsPrefix = result['ws']
    this.startConnect()
    window.addEventListener('message', this.handleEventFromLuna, false)
  },
  methods: {
    generateShareURL() {
      return `${BASE_URL}/lion/share/${this.shareId}/`
    },
    getMinuteLabel(item) {
      let minuteLabel = this.$t('Minute')
      if (item > 1) {
        minuteLabel = this.$t('Minutes')
      }
      return `${item} ${minuteLabel}`
    },
    shareDialogClosed() {
      this.$log.debug('share dialog closed')
      this.loading = false
      this.shareId = null
      this.shareCode = null
      this.shareDialogVisible = false
    },
    copyShareURL() {
      if (!this.shareId) {
        return
      }
      const shareURL = this.generateShareURL()
      this.$log.debug('share URL: ' + shareURL)
      const linkTitle = this.$t('LinkAddr')
      const codeTitle = this.$t('VerifyCode')
      const text = `${linkTitle}： ${shareURL}\n${codeTitle}: ${this.shareCode}`
      CopyTextToClipboard(text)
      this.$message(this.$t('CopyShareURLSuccess'))
    },
    getSessionUser(query) {
      if (query !== '') {
        this.userLoading = true
        getUserInfo(query).then(response => {
          this.userLoading = false
          this.$log.debug('User options response: ', response)
          this.userOptions = response
        })
      } else {
        this.userOptions = []
      }
    },
    handleShareURlCreated() {
      this.shareLoading = true
      const data = {
        session_id: this.session.id,
        expired_time: this.shareLinkRequest.expiredTime,
        users: this.shareLinkRequest.users,
        action_perm: this.shareLinkRequest.actionPerm
      }
      createShareURL(data).then(response => {
        this.$log.debug('Create share URL response: ', response)
        this.shareLoading = false
        this.shareId = response.id
        this.shareCode = response.verify_code
      })
      this.$log.debug('分享请求数据： ', this.sessionId, this.shareLinkRequest)
    },
    increaseScale() {
      this.autoFit = false
      this.scale += 0.1
      this.display.scale(this.scale)
    },
    decreaseScale() {
      this.autoFit = false
      this.scale -= 0.1
      if (this.scale < 0.5) {
        this.scale = 0.5
      }
      this.display.scale(this.scale)
    },
    toggleFileSystem(e) {
      if (this.menuDisable) {
        return
      }
      this.fileDrawer = !this.fileDrawer
    },
    initFileSystem() {
      this.fileSystemInited = true
      setTimeout(() => {
        const dropbox = document.getElementById('display')
        dropbox.addEventListener('dragenter', function(e) {
          e.stopPropagation()
          e.preventDefault()
        }, false)
        dropbox.addEventListener('dragover', function(e) {
          e.stopPropagation()
          e.preventDefault()
        }, false)
        dropbox.addEventListener('drop', this.$refs.fileSystem.fileDrop, false)
      }, 300)
    },
    initClipboard() {
      if (this.session && this.session.permission) {
        const actions = this.session.action_permission
        const hasClipboardPermission = actions.enable_copy || actions.enable_paste
        this.$log.debug(this.session.permission)
        this.clipboardInited = hasClipboardPermission
      }
      this.$log.debug('Clipboard inited: ', this.clipboardInited)
    },
    startConnect() {
      const params = getURLParams()
      const tokenId = params.get('token')
      this.getConnectString(tokenId).then(connectionParams => {
        this.connectGuacamole(connectionParams, this.wsPrefix)
      })
    },
    onClientConnected() {
      this.onWindowResize()
      window.addEventListener('resize', this.debounceWindowResize)
      window.onfocus = this.onWindowFocus
      window.onblur = () => {
        if (this.keyboard !== null) {
          this.keyboard.reset()
        }
      }
    },

    handleEventFromLuna(evt) {
      const msg = evt.data
      switch (msg.name) {
        case 'PING':
          if (this.lunaId != null) {
            return
          }
          this.lunaId = msg.id
          this.origin = evt.origin
          this.sendEventToLuna('PONG', null)
          break
        case 'OPEN':
          this.$refs.panel.toggle()
          break
      }
      console.log('Lion got post msg: ', msg)
    },

    sendEventToLuna(name, data) {
      if (this.lunaId != null) {
        window.parent.postMessage({ name: name, id: this.lunaId, data: data }, this.origin)
      }
    },

    menuIndex(index, num) {
      return index + num
    },
    toggleClipboard() {
      if (this.menuDisable) {
        return
      }
      this.clipboardDrawer = !this.clipboardDrawer
    },

    getAutoSize() {
      const width = window.innerWidth - sideWidth
      const height = window.innerHeight
      this.$log.debug('auto size:', width, height)
      return [width, height]
    },

    getGuaSize() {
      const lunaSetting = localStorageGet('LunaSetting') || {}
      const graphics = lunaSetting['graphics'] || {}
      const solution = graphics['rdp_resolution']
      if (!solution || solution.toLowerCase() === 'auto' || solution.indexOf('x') === -1) {
        this.$log.debug('Solution invalid: ', solution)
        return this.getAutoSize()
      }
      let [width, height] = solution.split('x')
      width = parseInt(width)
      height = parseInt(height)
      if (isNaN(width) || width < 100 || isNaN(height) || height < 100) {
        this.$log.debug('Solution invalid2: ', solution)
        return this.getAutoSize()
      }
      return [width, height]
    },
    getKeyboardLayout() {
      const lunaSetting = localStorageGet('LunaSetting') || {}
      const graphics = lunaSetting['graphics'] || {}
      const keyboardLayout = graphics['keyboard_layout']
      this.$log.debug('KeyboardLayout: ', keyboardLayout)
      if (!keyboardLayout) {
        return ''
      }
      return keyboardLayout
    },

    getConnectString(tokenId) {
      // Calculate optimal width/height for display
      const [optimalWidth, optimalHeight] = this.getGuaSize()
      const keyboardLayout = this.getKeyboardLayout()
      const optimalDpi = pixelDensity * 96
      return new Promise((resolve, reject) => {
        Promise.all([
          getSupportedMimetypes(),
          getSupportedGuacAudios(),
          getSupportedGuacVideos()
        ]).then(values => {
          // ["image/jpeg", "image/png", "image/webp"]
          const supportImages = values[0]
          const supportAudios = values[1]
          const supportVideos = values[2]
          let connectString =
              'TOKEN_ID=' + encodeURIComponent(tokenId) +
              '&GUAC_WIDTH=' + Math.floor(optimalWidth) +
              '&GUAC_HEIGHT=' + Math.floor(optimalHeight) +
              '&GUAC_DPI=' + Math.floor(optimalDpi) +
              '&GUAC_KEYBOARD=' + encodeURIComponent(keyboardLayout)
          this.$log.debug('Connect string: ', connectString)
          supportImages.forEach(function(mimetype) {
            connectString += '&GUAC_IMAGE=' + encodeURIComponent(mimetype)
          })
          supportAudios.forEach(function(mimetype) {
            connectString += '&GUAC_AUDIO=' + encodeURIComponent(mimetype)
          })
          supportVideos.forEach(function(mimetype) {
            connectString += '&GUAC_VIDEO=' + encodeURIComponent(mimetype)
          })
          resolve(connectString)
        })
      })
    },

    onTunnelStateChanged(state) {
      switch (state) {
        // Connection is being established
        case Guacamole.Tunnel.State.CONNECTING:
          this.tunnelState = 'CONNECTING'
          this.$log.debug('Tunnel state change to Tunnel.State.CONNECTING ')
          break

          // Connection is established / no longer unstable
        case Guacamole.Tunnel.State.OPEN:
          this.tunnelState = 'OPEN'
          this.$log.debug('Tunnel state change to Tunnel.State.OPEN ')
          this.initFileSystem()
          this.initClipboard()
          break

          // Connection is established but misbehaving
        case Guacamole.Tunnel.State.UNSTABLE:
          this.tunnelState = 'UNSTABLE'
          this.$log.debug('Tunnel state change to Tunnel.State.UNSTABLE ')
          break

          // Connection has closed
        case Guacamole.Tunnel.State.CLOSED:
          this.tunnelState = 'CLOSED'
          this.$log.debug('Tunnel state change to Tunnel.State.CLOSED ')
          break
        default:
          this.tunnelState = 'unknown'
          this.$log.debug('Tunnel state change tounknown ', state)
          break
      }
    },

    clientStateChanged(clientState) {
      switch (clientState) {
        // Idle
        case 0:
          this.clientState = 'IDLE'
          this.$log.debug('clientState, IDLE')
          break

          // Ignore "connecting" state
        case 1: // Connecting
          this.clientState = 'Connecting'
          this.$log.debug('clientState, Connecting')
          break

          // Connected + waiting
        case 2:
          this.clientState = 'Connected + waiting'
          this.$log.debug('clientState, Connected + waiting')
          break

          // Connected
        case 3:
          this.clientState = 'Connected'
          this.$log.debug('clientState, Connected ')
          this.loading = false
          // Send any clipboard data already provided
          // if (managedClient.clipboardData)
          //     ManagedClient.setClipboard(managedClient, managedClient.clipboardData);
          //
          // Begin streaming audio input if possible
          var AUDIO_INPUT_MIMETYPE = 'audio/L16;rate=44100,channels=2'
          var requestAudioStream = function requestAudioStream(client) {
            // Create new audio stream, associating it with an AudioRecorder
            const stream = client.createAudioStream(AUDIO_INPUT_MIMETYPE)
            let recorder
            try {
              recorder = Guacamole.AudioRecorder.getInstance(stream, AUDIO_INPUT_MIMETYPE)
            } catch (e) {
              console.log('Get audio recorder error, ignore')
              recorder = null
            }

            // If creation of the AudioRecorder failed, simply end the stream
            // eslint-disable-next-line brace-style
            if (!recorder) { stream.sendEnd() }
            // Otherwise, ensure that another audio stream is created after this
            // audio stream is closed
            else { recorder.onclose = requestAudioStream.bind(this, client) }
          }
          requestAudioStream(this.client)
          this.onClientConnected()
          break

          // Update history when disconnecting
        case 4: // Disconnecting
        case 5: // Disconnected
          this.clientState = 'Disconnecting'
          this.$log.debug('clientState, Disconnected ')
          // this.closeDisplay('clientState Disconnecting')
          var display = document.getElementById('display')
          display.innerHTML = ''
          this.sendEventToLuna('CLOSE', null)
          break
      }
    },

    clientOnErr(status) {
      this.loading = false
      this.$log.debug('clientOnErr', status)
      this.closeDisplay(status)
    },

    closeDisplay(status) {
      this.$log.debug(status, i18n.locale)
      const code = status.code
      let msg = status.message
      const currentLang = getLanguage()
      msg = ErrorStatusCodes[code] ? this.$t(ErrorStatusCodes[code]) : this.$t(ConvertGuacamoleError(status.message))
      switch (code) {
        case 1005:
          // 管理员终断会话，特殊处理
          if (currentLang === 'cn') {
            msg = status.message + ' ' + msg
          } else {
            msg = msg + ' ' + status.message
          }
          break
        case 1003:
          msg = msg.replace('{PLACEHOLDER}', status.message)
          break
        case 1010:
          msg = msg.replace('{PLACEHOLDER}', status.message)
          break
        case 1006:
          msg = msg + ': ' + status.message
          break
      }
      this.$alert(msg, this.$t('ErrTitle'), {
        confirmButtonText: this.$t('OK'),
        callback: action => {
          const display = document.getElementById('display')
          if (this.client) {
            // display.removeChild(this.client.getDisplay().getElement())
            display.innerHTML = ''
          }
        }
      })
    },
    onCursor(canvas, x, y) {
      this.localCursor = true
    },

    handleMouseState(mouseState) {
      // Do not attempt to handle mouse state changes if the client
      // or display are not yet available
      if (!this.client || !this.display) { return }
      this.sendEventToLuna('MOUSEEVENT', '')
      // Send mouse state, show cursor if necessary
      this.display.showCursor(!this.localCursor)
      this.sendScaledMouseState(mouseState)
    },

    sendScaledMouseState(mouseState) {
      const scaledState = new Guacamole.Mouse.State(
        mouseState.x / this.display.getScale(),
        mouseState.y / this.display.getScale(),
        mouseState.left,
        mouseState.middle,
        mouseState.right,
        mouseState.up,
        mouseState.down)
      this.client.sendMouseState(scaledState)
    },

    handleEmulatedMouseState(mouseState) {
      // Do not attempt to handle mouse state changes if the client
      // or display are not yet available
      if (!this.client || !this.display) { return }

      // Ensure software cursor is shown
      this.display.showCursor(true)
      this.$log.debug('handleEmulatedMouseState', mouseState)
      // Send mouse state, ensure cursor is visible
      this.sendScaledMouseState(mouseState)
    },

    handleEmulatedMouseDown(mouseState) {
      this.handleEmulatedMouseState(mouseState)
      this.isMenuCollapse = true
      this.sink.focus()
      this.$log.debug('handleEmulatedMouseDown', mouseState)
    },

    onMouseDown(mouseState) {
      document.body.focus()
      this.handleMouseState(mouseState)
      this.isMenuCollapse = true
      this.sink.focus()
    },

    onMouseOut(mouseState) {
      if (!this.display) return
      this.display.showCursor(false)
    },

    onCloseDrawer() {
      this.$log.debug('onCloseDrawer', this.sink)
      this.sink.focus()
    },

    getPropScale() {
      const display = this.client.getDisplay()
      if (!display) {
        return 1
      }
      const displayWidth = display.getWidth()
      const displayHeight = display.getHeight()
      if (displayWidth === 0 || displayHeight === 0) {
        return 1
      }
      const [width, height] = this.getAutoSize()
      // Calculate scale to fit screen
      return Math.min(
        width / Math.max(display.getWidth(), 1),
        height / Math.max(display.getHeight(), 1)
      )
    },

    updateDisplayScale() {
      const display = this.client.getDisplay()
      if (!display) {
        return
      }

      const scale = this.getPropScale()
      if (scale === this.scale) {
        return
      }
      this.scale = scale
      this.display.scale(scale)
    },

    debounceWindowResize() {
      if (this.timeout) {
        clearTimeout(this.timeout)
      }
      this.timeout = setTimeout(() => this.onWindowResize(), 300)
    },

    onWindowResize() {
      // 监听 window display的变化
      this.displayWidth = window.innerWidth - sideWidth
      this.displayHeight = window.innerHeight
      if (!this.autoFit) {
        this.$log.debug('onWindowResize, autoFit is false')
        return
      }
      const [optimalWidth, optimalHeight] = this.getGuaSize()
      if (this.client !== null) {
        const display = this.client.getDisplay()
        const displayHeight = display.getHeight() * pixelDensity
        const displayWidth = display.getWidth() * pixelDensity
        this.updateDisplayScale()
        if (displayHeight === optimalWidth && displayWidth === optimalHeight) {
          return
        }
        this.client.sendSize(optimalWidth, optimalHeight)
      }
    },

    displayResize(width, height) {
      // 监听guacamole display的变化
      this.$log.debug('Display resize: ', width, height)
      this.updateDisplayScale()
    },

    onWindowFocus() {
      this.$log.debug('On window focus ')
      if (this.$refs.clipboard) {
        this.$refs.clipboard.sendClipboardToRemote()
      }
    },

    onsync: function(timestamp) {
      // this.$log.debug('onsync==> ', timestamp)
    },

    handleKeys(keys) {
      if (!this.client) {
        return
      }
      for (let i = 0; i < keys.length; i++) {
        this.client.sendKeyEvent(1, keys[i])
      }
      for (let i = 0; i < keys.length; i++) {
        this.client.sendKeyEvent(0, keys[i])
      }
    },

    setTunnelCallback(tunnel) {
      const vm = this
      tunnel.onerror = (status) => {
        vm.$message.error(vm.$t('WebSocketError'))
        vm.$log.error('Tunnel error: ', status)
      }
      tunnel.onuuid = (uuid) => {
        vm.$log.debug('Tunnel assigned UUID: ', uuid)
        vm.tunnel.uuid = uuid
      }
      tunnel.onstatechange = vm.onTunnelStateChanged
      const oninstruction = tunnel.oninstruction
      tunnel.oninstruction = (opcode, argv) => {
        if (oninstruction) {
          oninstruction(opcode, argv)
        }
        if (opcode === 'jms_event') {
          vm.onJmsEvent(argv[0], argv[1])
          vm.$log.debug('Tunnel instruction: ', opcode, argv)
        }
      }
    },

    onJmsEvent(event, data) {
      const dataObj = JSON.parse(data)
      switch (event) {
        case 'session_pause': {
          const msg = `${dataObj.user} ${this.$t('PauseSession')}`
          this.$message.info(msg)
          break
        }
        case 'session_resume': {
          const msg = `${dataObj.user} ${this.$t('ResumeSession')}`
          this.$message.info(msg)
          break
        }
        case 'session': {
          this.session = dataObj
          this.initFileSystem()
          this.initClipboard()
          const actions = this.session.action_permission
          this.$log.debug('Session actions: ', actions)
          this.enableShare = actions.enable_share
          break
        }
        case 'current_user': {
          this.shareid = dataObj.share_id
          break
        }
        case 'share_join': {
          if (dataObj.primary) {
            this.$log.debug('primary user 不提醒')
            break
          }
          const joinMsg = `${dataObj.user} ${this.$t('JoinShare')}`
          this.$message(joinMsg)
          break
        }
        case 'share_exit': {
          const leaveMsg = `${dataObj.user} ${this.$t('LeaveShare')}`
          this.$message(leaveMsg)
          break
        }
        case 'share_users': {
          this.onlineUsersMap = dataObj
          break
        }
        default:
          break
      }
      this.$log.debug('onJmsEvent: ', event, data)
    },

    onFileSystem(obj, name) {
      this.hasFileSystem = true
      if (this.$refs.fileSystem) { this.$refs.fileSystem.fileSystemReceived(obj, name) }
    },

    setClientCallback(client) {
      const vm = this
      // client.onrequired = this.onRequireParams
      client.onstatechange = this.clientStateChanged
      client.onerror = this.clientOnErr
      // 文件挂载
      client.onfilesystem = this.onFileSystem
      client.onfile = (stream, mimetype, filename) => {
        return vm.$refs.fileSystem.clientFileReceived(stream, mimetype, filename)
      }
      // 剪贴板
      client.onclipboard = (stream, mimetype) => {
        return vm.$refs.clipboard.receiveClientClipboard(stream, mimetype)
      }
      client.onsync = this.onsync
    },
    setDisplayCallback(display) {
      display.onresize = this.displayResize
      display.oncursor = this.onCursor

      const displayEl = display.getElement()
      this.$log.debug('Display el: ', displayEl)
      displayEl.onclick = (e) => {
        e.preventDefault()
        return false
      }
      const mouse = new Guacamole.Mouse(displayEl)
      // Ensure focus is regained via mousedown before forwarding event
      mouse.onmousedown = this.onMouseDown
      mouse.onmouseup = mouse.onmousemove = this.handleMouseState
      // Hide software cursor when mouse leaves display
      mouse.onmouseout = this.onMouseOut
      this.mouse = mouse
      // touch 触屏操作
      const outDisplay = document.getElementById('display')
      const touchScreen = new Guacamole.Mouse.Touchscreen(outDisplay)
      touchScreen.onmousedown = this.handleEmulatedMouseDown
      touchScreen.onmousemove = touchScreen.onmouseup = this.handleEmulatedMouseState
      // 输入下沉
      const sink = new Guacamole.InputSink()
      sink.focus()
      this.sink = sink

      // Keyboard
      const keyboard = new Guacamole.Keyboard(sink.getElement())
      keyboard.onkeydown = (keysym) => {
        this.sendEventToLuna('KEYBOARDEVENT', '')
        this.client.sendKeyEvent(1, keysym)
      }
      keyboard.onkeyup = (keysym) => {
        this.sendEventToLuna('KEYBOARDEVENT', '')
        this.client.sendKeyEvent(0, keysym)
      }
      this.keyboard = keyboard
    },
    connectGuacamole(connectionParams, wsURL) {
      const displayRef = document.getElementById('display')
      const tunnel = new Guacamole.WebSocketTunnel(wsURL)
      // 连接资产耗时，造成的 ws 超时断开问题 默认 15s 改成 60s
      tunnel.receiveTimeout = 60 * 1000
      const client = new Guacamole.Client(tunnel)
      this.client = client
      this.tunnel = tunnel
      this.display = client.getDisplay()

      this.setTunnelCallback(this.tunnel)
      this.setClientCallback(this.client)
      this.setDisplayCallback(this.display)
      // client.getDisplay()
      displayRef.appendChild(this.display.getElement())
      displayRef.appendChild(this.sink.getElement())
      this.$log.debug('Display : ', displayRef)
      // 开始连接
      client.connect(connectionParams)
      window.onunload = function() {
        client.disconnect()
      }
    }
  }
}
</script>

<style scoped>
.el-container {
  margin: 0 auto;
}

.el-main {
  padding: 0;
  overflow: hidden;
  width: 100%;
  height: 100%;
}
.el-row {
  height: 100%;
}

.el-aside {
  height: 100%;
  overflow: hidden;
}

.el-dropdown-link {
  color: #409eff;
}

.el-icon-arrow-down {
  font-size: 12px;
}
.el-menu {
  height: 100%;
  min-width: 100%;
  border-right: 0 none;
}

.el-menu-item {
  color: rgba(0, 0, 0, 0.65);
  font-size: 14px;
  list-style-type: none;
  cursor: pointer;
  border-radius: 2px;
  padding: 0;
}

.el-menu-item:hover,.el-menu-item:focus{
  color: white;
  background: rgba(0, 0, 0, .3);
}

#displayOuter {
  height: 100%;
  width: 100%;
  position: absolute;
  left: 0;
  top: 0;
  display: table;
}

#displayMiddle {
  width: 100%;
  height: 100%;
  display: table-cell;
  vertical-align: middle;
  text-align: center;
}

#display{
  display: inline-block;
}

#display * {
  position: relative;
}

#display > * {
  margin-left: auto;
  margin-right: auto;
}

.sidebar-popper .el-menu--popup{
  background-color: white;
  color: #1f1b1b;
  padding: 0;
}

.h3title{
  padding-left: 25px;
}

.share-dialog >>> .el-dialog__header {
  background: #333;
}

.share-dialog >>> .el-dialog__body {
  background: #333;
}

.share-dialog >>> .el-dialog__title {
  color: #ffffff;
}

.share-dialog >>> .el-form-item__label {
  color: #ffffff;
  background: #303133;
}

.share-dialog >>> .el-button:hover {
  background: rgb(134, 133, 133);
  color: #303133;
  border-color: rgba(183, 172, 172, 0.1);
}
.share-btn >>> .el-button {
  background: #303133;
  color: #ffffff;
}

.share-btn:hover {
  background: rgb(134, 133, 133);
  color: #303133;
  border-color: rgba(183, 172, 172, 0.1);
}

.share >>> .el-button {
  background: #303133;
  color: #ffffff;
}

.copy-btn:hover {
  background: rgb(134, 133, 133);
  color: #303133;
  border-color: rgba(183, 172, 172, 0.1);
}
</style>
