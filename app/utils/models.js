import request from '@/utils/request'

const CACHE_TTL = 60 * 60 * 1000

const cache = {}

export async function getModel(model_id) {
  const now = Date.now()

  if (cache[model_id]) {
    if (now - cache[model_id].timestamp < CACHE_TTL) {
      return cache[model_id].data
    }
    delete cache[model_id]
  }

  const res = await request({
    url: `/model/${model_id}`,
    method: 'get'
  })
  const data = res.data || res

  const cacheItem = {
    timestamp: now,
    data: data
  }
  cache[model_id] = cacheItem

  return data
}

export function clearCache(model_id) {
  delete cache[model_id]
}

export function clearAllCache() {
  Object.keys(cache).forEach(k => delete cache[k])
}
