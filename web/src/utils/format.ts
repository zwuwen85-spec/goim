export const parseSqlNullString = (val: any): string => {
  if (!val) return ''
  if (typeof val === 'string') return val
  if (typeof val === 'object' && 'String' in val) {
    return val.Valid ? val.String : ''
  }
  return String(val)
}

export const getAvatarUrl = (val: any): string => {
  return parseSqlNullString(val)
}
