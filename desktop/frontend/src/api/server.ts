// 아지트·방·메시지·짤 서버 API. Go(APIService.Call)가 토큰을 붙여 서버로 보낸다.
// 모양은 proto/*.go 와 같다 (JSON 이름 그대로).
import { APIService } from '../../bindings/github.com/TogetherBros/azit/desktop/internal/bridge'

type Method = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'

async function call<T>(method: Method, path: string, body?: unknown): Promise<T> {
  return (await APIService.Call(method, path, body ?? null)) as T
}

/* ---------- 모양 (proto) ---------- */

export type Role = 'super_admin' | 'admin' | 'member'

export type AzitRes = {
  azit_id: number
  name: string
  color: string
  description: string // 한 줄 소개
  banner_url: string
  member_count: number
  my_role: Role | ''
  my_nickname: string
}

export type InviteRes = {
  code: string // XXXX-XXXX
  azit_id: number
  created_by: number
  expires_at: string | null
  max_uses: number // 0 = 무제한
  uses: number
  created_at: string
}

export type InvitePreview = {
  code: string
  azit: AzitRes
  inviter_nickname: string
  already_member: boolean
}

export type ExpiresIn = '30m' | '1h' | '1d' | '7d' | 'never'

export type RoomRes = { room_id: number; azit_id: number; name: string; position: number; unread: number }

export type SenderRes = { account_id: number; nickname: string; color: string; avatar_url: string }

export type ReactionRes = { text: string; meme_id: number; meme_url: string; count: number; mine: boolean }

export type MessageRes = {
  message_id: number
  room_id: number
  temp_id: string
  kind: 'text' | 'photo' | 'meme' | 'system'
  body: string // photo, meme 이면 ""
  meme_id: number // meme 이면 짤 id
  media_url: string // photo, meme 이면 그림 보기 링크
  reply_to_id: number | null
  shake_ids: number[]
  shake_all: boolean
  created_at: string
  edited_at: string | null
  deleted: boolean
  sender: SenderRes | null
  reactions: ReactionRes[]
}

export type MessageList = { messages: MessageRes[]; has_more: boolean; last_read_id: number }

export type MemeRes = { meme_id: number; name: string; url: string; created_by: number; created_at: string }

export type ReactRes = { react_id: number; text: string; meme_id: number; meme_url: string; position: number }

// 맞장구 하나 (글자 또는 짤). 달기·떼기·목록 칸 공통
export type ReactItem = { text: string; meme_id: number }

export type MemberRes = {
  account_id: number; name: string; nickname: string; color: string; avatar_url: string; banner_url: string; bio: string; role: Role; joined_at: string
}

export type NotificationRes = {
  notification_id: number
  kind: 'shake' | 'join' | 'kick'
  azit_id: number
  azit_name: string
  room_id: number // shake 면 그 방
  room_name: string
  message_id: number
  actor_id: number
  actor_nickname: string
  preview: string // shake 면 메시지 앞부분
  read: boolean
  created_at: string
}

/* ---------- 알림 ---------- */

export const listNotifications = () =>
  call<{ notifications: NotificationRes[]; unread: number }>('GET', '/v1/notifications')

export const readAllNotifications = () => call<null>('POST', '/v1/notifications/read-all')

/* ---------- 아지트 · 초대 ---------- */

export const listAzits = () => call<AzitRes[]>('GET', '/v1/azits')

export const createAzit = (name: string, color: string, bannerKey = '') =>
  call<AzitRes>('POST', '/v1/azits', { name, color, banner_key: bannerKey })

// 기본 정보 바꾸기 (관리자 이상). 바꿀 칸만 보낸다. banner_key 가 "" 면 배너를 지운다
export type AzitPatch = { name?: string; color?: string; description?: string; banner_key?: string }

export const updateAzit = (azitId: number, patch: AzitPatch) => call<AzitRes>('PATCH', `/v1/azits/${azitId}`, patch)

// 코드는 하이픈이 있어도 없어도 된다 (서버가 정리한다)
export const previewInvite = (code: string) => call<InvitePreview>('GET', `/v1/invites/${encodeURIComponent(code)}`)

export const joinInvite = (code: string) => call<AzitRes>('POST', `/v1/invites/${encodeURIComponent(code)}/join`)

export const listInvites = (azitId: number) => call<InviteRes[]>('GET', `/v1/azits/${azitId}/invites`)

export const createInvite = (azitId: number, expiresIn: ExpiresIn, maxUses: number) =>
  call<InviteRes>('POST', `/v1/azits/${azitId}/invites`, { expires_in: expiresIn, max_uses: maxUses })

export const revokeInvite = (code: string) => call<null>('DELETE', `/v1/invites/${encodeURIComponent(code)}`)

/* ---------- 방 ---------- */

export const listRooms = (azitId: number) => call<RoomRes[]>('GET', `/v1/azits/${azitId}/rooms`)

export const createRoom = (azitId: number, name: string) => call<RoomRes>('POST', `/v1/azits/${azitId}/rooms`, { name })

export const renameRoom = (azitId: number, roomId: number, name: string) =>
  call<RoomRes>('PATCH', `/v1/azits/${azitId}/rooms/${roomId}`, { name })

export const reorderRooms = (azitId: number, roomIds: number[]) =>
  call<RoomRes[]>('PUT', `/v1/azits/${azitId}/rooms/order`, { room_ids: roomIds })

/* ---------- 메시지 ---------- */

export const listMessages = (azitId: number, roomId: number, before?: number, limit = 50) =>
  call<MessageList>('GET', `/v1/azits/${azitId}/rooms/${roomId}/messages?limit=${limit}${before ? `&before=${before}` : ''}`)

// 글자: body · 사진: kind 'photo' + photo_key (uploadFile('photo') 로 받은 키) · 짤: kind 'meme' + meme_id
export type SendInput = {
  temp_id: string
  kind?: 'text' | 'photo' | 'meme'
  body?: string
  photo_key?: string
  meme_id?: number
  shake_ids?: number[]
  shake_all?: boolean
}

export const sendMessage = (azitId: number, roomId: number, input: SendInput) =>
  call<MessageRes>('POST', `/v1/azits/${azitId}/rooms/${roomId}/messages`, input)

export const markRead = (azitId: number, roomId: number, messageId: number) =>
  call<null>('POST', `/v1/azits/${azitId}/rooms/${roomId}/read`, { message_id: messageId })

export const editMessage = (azitId: number, messageId: number, body: string) =>
  call<MessageRes>('PATCH', `/v1/azits/${azitId}/messages/${messageId}`, { body })

export const deleteMessage = (azitId: number, messageId: number) =>
  call<null>('DELETE', `/v1/azits/${azitId}/messages/${messageId}`)

export const addReaction = (azitId: number, messageId: number, r: ReactItem) =>
  call<ReactionRes[]>('PUT', `/v1/azits/${azitId}/messages/${messageId}/reactions`, r)

export const removeReaction = (azitId: number, messageId: number, r: ReactItem) =>
  call<ReactionRes[]>('DELETE', `/v1/azits/${azitId}/messages/${messageId}/reactions`, r)

/* ---------- 짤 · 맞장구 목록 ---------- */

// 멤버
export const listMembers = (azitId: number) => call<MemberRes[]>('GET', `/v1/azits/${azitId}/members`)

export const updateMyNickname = (azitId: number, nickname: string) =>
  call<MemberRes>('PATCH', `/v1/azits/${azitId}/members/me`, { nickname })

export const leaveAzit = (azitId: number) => call<null>('DELETE', `/v1/azits/${azitId}/members/me`)

export const updateMemberRole = (azitId: number, accountId: number, role: 'admin' | 'member') =>
  call<MemberRes>('PATCH', `/v1/azits/${azitId}/members/${accountId}/role`, { role })

export const kickMember = (azitId: number, accountId: number) => call<null>('DELETE', `/v1/azits/${azitId}/members/${accountId}`)

export const listMemes = (azitId: number) => call<MemeRes[]>('GET', `/v1/azits/${azitId}/memes`)

export const createMeme = (azitId: number, name: string, key: string) =>
  call<MemeRes>('POST', `/v1/azits/${azitId}/memes`, { name, key })

export const deleteMeme = (azitId: number, memeId: number) => call<null>('DELETE', `/v1/azits/${azitId}/memes/${memeId}`)

export const listReacts = (azitId: number) => call<ReactRes[]>('GET', `/v1/azits/${azitId}/reacts`)

export const replaceReacts = (azitId: number, reacts: ReactItem[]) =>
  call<ReactRes[]>('PUT', `/v1/azits/${azitId}/reacts`, { reacts })

// 사진을 R2 에 올리고 key 를 받는다 (짤, 채팅 사진, 아지트 배너)
export const uploadFile = (kind: 'meme' | 'photo' | 'azit_banner',contentType: string, base64: string) =>
  APIService.Upload(kind, contentType, base64)

/* ---------- 오류 ---------- */

// Go 에서 온 오류를 화면에 띄울 문장으로 바꾼다. 서버 오류 문장은 이미 한국어라 그대로 쓴다.
export function errorText(e: unknown): string {
  const msg = e instanceof Error ? e.message : String(e)
  if (/refused|dial tcp|connect/i.test(msg)) return '서버에 연결하지 못했어요. 서버가 켜져 있는지 봐 주세요'
  if (/[가-힣]/.test(msg)) return msg
  return '잠시 후 다시 해 주세요'
}

export const isAdmin = (role: Role | '') => role === 'admin' || role === 'super_admin'
