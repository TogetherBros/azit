// 고른 사진을 올리기 좋게 줄인다. 원본이 커도 프사·배너로는 이 정도면 충분하다.
// 프사: 가운데를 정사각형으로 잘라 512px. 배너: 비율 그대로 긴 쪽 1500px.

export type ShrunkImage = {
  contentType: string // 브라우저가 실제로 만든 형식 (webp 를 못 만들면 png 로 나옴)
  base64: string // Go 로 넘길 바이트
  size: number
}

const AVATAR_SIDE = 512
const BANNER_LONG = 1500

export async function shrinkImage(file: File, kind: 'avatar' | 'banner'): Promise<ShrunkImage> {
  const img = await loadImage(file)

  let sx = 0, sy = 0, sw = img.naturalWidth, sh = img.naturalHeight
  let dw: number, dh: number

  if (kind === 'avatar') {
    const side = Math.min(sw, sh)
    sx = (sw - side) / 2
    sy = (sh - side) / 2
    sw = sh = side
    dw = dh = Math.min(side, AVATAR_SIDE)
  } else {
    const scale = Math.min(1, BANNER_LONG / Math.max(sw, sh))
    dw = Math.round(sw * scale)
    dh = Math.round(sh * scale)
  }

  const canvas = document.createElement('canvas')
  canvas.width = dw
  canvas.height = dh
  const ctx = canvas.getContext('2d')
  if (!ctx) throw new Error('사진을 줄이지 못했어요')
  ctx.imageSmoothingQuality = 'high'
  ctx.drawImage(img, sx, sy, sw, sh, 0, 0, dw, dh)

  const blob = await new Promise<Blob | null>((ok) => canvas.toBlob(ok, 'image/webp', 0.9))
  if (!blob) throw new Error('사진을 줄이지 못했어요')

  return { contentType: blob.type, base64: await toBase64(blob), size: blob.size }
}

function loadImage(file: File): Promise<HTMLImageElement> {
  return new Promise((ok, fail) => {
    const url = URL.createObjectURL(file)
    const img = new Image()
    img.onload = () => { URL.revokeObjectURL(url); ok(img) }
    img.onerror = () => { URL.revokeObjectURL(url); fail(new Error('이 사진은 열 수 없어요')) }
    img.src = url
  })
}

async function toBase64(blob: Blob): Promise<string> {
  const bytes = new Uint8Array(await blob.arrayBuffer())
  let bin = ''
  for (let i = 0; i < bytes.length; i += 0x8000) {
    bin += String.fromCharCode(...bytes.subarray(i, i + 0x8000))
  }
  return btoa(bin)
}
