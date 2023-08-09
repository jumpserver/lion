export const ErrorStatusCodes = {
  256: 'GuaErrUnSupport',
  519: 'GuaErrUpstreamNotFound',
  514: 'GuaErrUpStreamTimeout',
  521: 'GuaErrSessionConflict',
  769: 'GuaErrClientUnauthorized',
  1000: 'JMSErrNoSession',
  1001: 'JMSErrAuthUser',
  1002: 'JMSErrBadParams',
  1003: 'JMSErrIdleTimeOut',
  1004: 'JMSErrPermissionExpired',
  1005: 'JMSErrTerminatedByAdmin',
  1006: 'JMSErrAPIFailed',
  1007: 'JMSErrGatewayFailed',
  1008: 'JMSErrGuacamoleServer',
  1009: 'JMSErrDisconnected',
  1010: 'JMSErrMaxSession'
}

export function ConvertAPIError(errMsg) {
  if (typeof errMsg !== 'string') {
    return errMsg
  }
  const errArray = errMsg.split(':')
  if (errArray.length >= 1) {
    return APIErrorType[errArray[0]] || errMsg
  }
  return errMsg
}

export const APIErrorType = {
  'connect API core err': 'JMSErrAPIFailed',
  'unsupported type': 'JMSErrBadParams',
  'unsupported protocol': 'JMSErrBadParams',
  'permission deny': 'JMSErrPermission'
}

export function ConvertGuacamoleError(errMsg) {
  if (typeof errMsg !== 'string') {
    return errMsg
  }
  return GuacamoleErrMsg[errMsg] || errMsg
}

export const GuacamoleErrMsg = {
  'Disconnected.': 'GuacamoleErrDisconnected',
  'Credentials expired.': 'GuacamoleErrCredentialsExpired',
  'Security negotiation failed (wrong security type?)': 'GuacamoleErrSecurityNegotiationFailed',
  'Access denied by server (account locked/disabled?)': 'GuacamoleErrAccessDenied',
  'Authentication failure (invalid credentials?)': 'GuacamoleErrAuthenticationFailure',
  'SSL/TLS connection failed (untrusted/self-signed certificate?)': 'GuacamoleErrSSLTLSConnectionFailed',
  'DNS lookup failed (incorrect hostname?)': 'GuacamoleErrDNSLookupFailed',
  'Server refused connection (wrong security type?)': 'GuacamoleErrServerRefusedConnectionBySecurityType',
  'Connection failed (server unreachable?)': 'GuacamoleErrConnectionFailed',
  'Upstream error.': 'GuacamoleErrUpstreamError',
  'Forcibly disconnected.': 'GuacamoleErrForciblyDisconnected',
  'Logged off.': 'GuacamoleErrLoggedOff',
  'Idle session time limit exceeded.': 'GuacamoleErrIdleSessionTimeLimitExceeded',
  'Active session time limit exceeded.': 'GuacamoleErrActiveSessionTimeLimitExceeded',
  'Disconnected by other connection.': 'GuacamoleErrDisconnectedByOtherConnection',
  'Server refused connection.': 'GuacamoleErrServerRefusedConnection',
  'Insufficient privileges.': 'GuacamoleErrInsufficientPrivileges',
  'Manually disconnected.': 'GuacamoleErrManuallyDisconnected',
  'Manually logged off.': 'GuacamoleErrManuallyLoggedOff',

  'Unsupported credential type requested.': 'GuacamoleErrUnsupportedCredentialTypeRequested'
}
