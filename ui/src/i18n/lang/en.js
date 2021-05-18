import el from 'element-ui/lib/locale/lang/en'

const message = {
  Shortcuts: 'Shortcuts',
  Clipboard: 'Clipboard',
  Files: 'Files',
  UploadFile: 'Upload File',
  ClearDone: 'Clear done',
  Connecting: 'Connecting',
  ErrTitle: 'Connect Error',
  JMSErrNoSession: 'Not found session',
  JMSErrAuthUser: 'Not auth user',
  JMSErrBadParams: 'Bad request params',
  JMSErrIdleTimeOut: 'Terminated by idle timeout',
  JMSErrPermissionExpired: 'Terminated by permission expired',
  JMSErrTerminatedByAdmin: 'Terminated by Admin',
  JMSErrAPIFailed: 'Core API failed',
  JMSErrGatewayFailed: 'Gateway not available',
  JMSErrGuacamoleServer: 'Connect guacamole server failed',
  GuaErrUpstreamNotFound: 'The remote desktop server does not appear to exist, or cannot be reached over the network.',
  GuaErrSessionConflict: 'The session has ended because it conflicts with another session.',
  GuaErrClientUnauthorized: 'User failed to logged in. (username and password are incorrect)',
  GuaErrUnSupport: 'The requested operation is unsupported.',
  GuaErrUpStreamTimeout: 'The remote desktop server is not responding.',
  OK: 'Ok',
  Submit: 'Submit',
  Cancel: 'Cancel',
  Skip: 'Skip',
  Username: 'Username',
  Password: 'Password',
  RequireParams: 'Require Params'
}

export default {
  ...el,
  ...message
}

