
export function ObjectDeepCompare(obj1: any, obj2: any): any {
  const keys1 = Object.keys(obj1);
  const keys2 = Object.keys(obj2);

  // 检查属性数量是否一致
  if (keys1.length !== keys2.length) {
    return false;
  }

  for (const key of keys1) {
    const val1 = obj1[key];
    const val2 = obj2[key];

    // 检查属性是否存在
    if (!(key in obj2)) {
      return false;
    }

    // 检查属性值是否相等
    if (val1 !== val2) {
      // 如果属性值不相等且都是对象，则递归比较
      if (typeof val1 === 'object' && typeof val2 === 'object' && val1 !== null && val2 !== null) {
        if (!ObjectDeepCompare(val1, val2)) {
          return false;
        }
      } else {
        return false;
      }
    }
  }

  return true;
}


export function LinkReplaceParams(link: string, data: any): string {
  if (!data) return link //无参数，直接返回

  //link.matchAll(/:\w+/g)
  let match = link.match(/\:\w+/g)
  if (match != null) {
    match.forEach(m => {
      const k = m.substring(1)
      const v = encodeURIComponent(data[k])
      link = link.replaceAll(m, v)
    })
  }
  return link
}


