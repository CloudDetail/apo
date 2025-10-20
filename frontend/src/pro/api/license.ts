import { get, post, headers } from 'src/core/utils/request'

interface uploadLicenseParams {
  output?: string;
  licenseFile?: string;
}

export const uploadLicense = async (params: uploadLicenseParams) => {
  return post('/api/license/upload', params, headers.formUrlencoded)
}

export const getLicense = async () => {
  return get('/api/license/current')
}