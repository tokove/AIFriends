const USERNAME_MIN_LENGTH = 2
const USERNAME_MAX_LENGTH = 32
const PASSWORD_MIN_LENGTH = 8
const PASSWORD_MAX_LENGTH = 72
const CHARACTER_NAME_MIN_LENGTH = 2
const CHARACTER_NAME_MAX_LENGTH = 32
const CHARACTER_PROFILE_MAX_LENGTH = 100000

function getLength(value) {
  return Array.from(value ?? '').length
}

export function validateUsername(value) {
  const username = value?.trim() ?? ''
  const length = getLength(username)

  if (!username) {
    return '用户名不能为空'
  }
  if (length < USERNAME_MIN_LENGTH || length > USERNAME_MAX_LENGTH) {
    return `用户名长度需在 ${USERNAME_MIN_LENGTH}-${USERNAME_MAX_LENGTH} 个字符之间`
  }
  return ''
}

export function validatePassword(value) {
  const password = value ?? ''
  const length = getLength(password)

  if (!password.trim()) {
    return '密码不能为空'
  }
  if (length < PASSWORD_MIN_LENGTH || length > PASSWORD_MAX_LENGTH) {
    return `密码长度需在 ${PASSWORD_MIN_LENGTH}-${PASSWORD_MAX_LENGTH} 个字符之间`
  }
  return ''
}

export function validateCharacterName(value) {
  const name = value?.trim() ?? ''
  const length = getLength(name)

  if (!name) {
    return '名字不能为空'
  }
  if (length < CHARACTER_NAME_MIN_LENGTH || length > CHARACTER_NAME_MAX_LENGTH) {
    return `名字长度需在 ${CHARACTER_NAME_MIN_LENGTH}-${CHARACTER_NAME_MAX_LENGTH} 个字符之间`
  }
  return ''
}

export function validateCharacterProfile(value) {
  const profile = value?.trim() ?? ''
  const length = getLength(profile)

  if (!profile) {
    return '角色介绍不能为空'
  }
  if (length > CHARACTER_PROFILE_MAX_LENGTH) {
    return `介绍太长了，最多支持 ${CHARACTER_PROFILE_MAX_LENGTH} 个字符`
  }
  return ''
}

export const formRules = {
  usernameMaxLength: USERNAME_MAX_LENGTH,
  passwordMaxLength: PASSWORD_MAX_LENGTH,
  characterNameMaxLength: CHARACTER_NAME_MAX_LENGTH,
  characterProfileMaxLength: CHARACTER_PROFILE_MAX_LENGTH,
}
