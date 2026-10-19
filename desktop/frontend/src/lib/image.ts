// 고른 사진을 올리기 좋게 줄인다. 원본이 커도 프사·배너로는 이 정도면 충분하다.
// 프사: 가운데를 정사각형으로 잘라 512px. 배너: 가운데를 12:5 로 잘라 1440x600.
// 배너를 화면 창문과 같은 비율로 미리 잘라 둬야 어디서 보든 똑같이 보인다.

export type ShrunkImage = {
  contentType: string // 실제로 만든 형식 (webp, 안 되면 jpeg)
  base64: string // Go 로 넘길 바이트
  size: number
}

const AVATAR_SIDE = 512
const BANNER_W = 1440
const BANNER_H = 600

const MEME_SIDE = 512
const PHOTO_SIDE = 1600

// avatar: 정사각형 자르기, banner·azit_banner: 12:5 자르기,
// meme·photo: 자르지 않고 긴 변을 줄이기 (짤 512px, 채팅 사진 1600px)
export async function shrinkImage(file: File, kind: 'avatar' | 'banner' | 'azit_banner' | 'meme' | 'photo'): Promise<ShrunkImage> {
  const img = await loadImage(file)

  let sx = 0, sy = 0, sw = img.naturalWidth, sh = img.naturalHeight
  let dw: number, dh: number

  if (kind === 'meme' || kind === 'photo') {
    const scale = Math.min(1, (kind === 'meme' ? MEME_SIDE : PHOTO_SIDE) / Math.max(sw, sh))
    dw = Math.round(sw * scale)
    dh = Math.round(sh * scale)
  } else if (kind === 'avatar') {
    const side = Math.min(sw, sh)
    sx = (sw - side) / 2
    sy = (sh - side) / 2
    sw = sh = side
    dw = dh = Math.min(side, AVATAR_SIDE)
  } else {
    const ratio = BANNER_W / BANNER_H
    if (sw / sh > ratio) {
      const w = sh * ratio // 옆으로 길면 양옆을 자른다
      sx = (sw - w) / 2
      sw = w
    } else {
      const h = sw / ratio // 위아래로 길면 위아래를 자른다
      sy = (sh - h) / 2
      sh = h
    }
    dw = Math.round(Math.min(sw, BANNER_W))
    dh = Math.round(dw / ratio)
  }

  const canvas = document.createElement('canvas')
  canvas.width = dw
  canvas.height = dh
  const ctx = canvas.getContext('2d')
  if (!ctx) throw new Error('사진을 줄이지 못했어요')
  ctx.imageSmoothingQuality = 'high'
  ctx.drawImage(img, sx, sy, sw, sh, 0, 0, dw, dh)

  // 맥(WebKit)은 webp 로 못 만들고 png 를 돌려준다. png 는 사진치고 너무 커서 그땐 jpeg 로 다시 만든다.
  let blob = await toBlob(canvas, 'image/webp')
  if (!blob || blob.type !== 'image/webp') blob = await toBlob(canvas, 'image/jpeg')
  if (!blob) throw new Error('사진을 줄이지 못했어요')

  return { contentType: blob.type, base64: await toBase64(blob), size: blob.size }
}

const MEME_MAX_BYTES = 10 << 20 // 서버 짤 업로드 한도 (storage/key.go 와 같게)

// 짤 올리기 준비. 움직이는 GIF·WebP 는 캔버스로 줄이면 첫 장면만 남아서 그대로 올린다 (10MB 까지).
// 움직이지 않는 사진은 shrinkImage 로 긴 변 512px 로 줄인다.
export async function prepareMeme(file: File): Promise<ShrunkImage> {
  const bytes = new Uint8Array(await file.arrayBuffer())
  if (isAnimated(file.type, bytes)) {
    if (bytes.length > MEME_MAX_BYTES) throw new Error('움직이는 짤은 10MB까지 올릴 수 있어요')
    return { contentType: file.type, base64: bytesToBase64(bytes), size: bytes.length }
  }
  return shrinkImage(file, 'meme')
}

// GIF 는 항상 그대로 (한 장짜리 GIF 도 줄일 이득이 적다). WebP 는 RIFF 안에 ANIM 조각이 있으면 움직이는 것.
function isAnimated(type: string, b: Uint8Array): boolean {
  if (type === 'image/gif') return true
  if (type !== 'image/webp' || b.length < 16) return false
  const ascii = (i: number, n: number) => String.fromCharCode(...b.subarray(i, i + n))
  if (ascii(0, 4) !== 'RIFF' || ascii(8, 4) !== 'WEBP') return false
  // VP8X 머리의 flags 에 애니메이션 비트(0x02)가 있다
  return ascii(12, 4) === 'VP8X' && b.length > 20 && (b[20] & 0x02) !== 0
}

function bytesToBase64(bytes: Uint8Array): string {
  let bin = ''
  for (let i = 0; i < bytes.length; i += 0x8000) {
    bin += String.fromCharCode(...bytes.subarray(i, i + 0x8000))
  }
  return btoa(bin)
}

function toBlob(canvas: HTMLCanvasElement, type: string): Promise<Blob | null> {
  return new Promise((ok) => canvas.toBlob(ok, type, 0.9))
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
  return bytesToBase64(new Uint8Array(await blob.arrayBuffer()))
}
