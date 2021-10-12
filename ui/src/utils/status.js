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
  1008: 'JMSErrGuacamoleServer'
}

export function ConvertAPIError(errMsg) {
  if (typeof errMsg !== 'string') {
    return errMsg
  }
  const errArray = errMsg.split(':')
  if (errArray.length > 1) {
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
