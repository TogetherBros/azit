// 내 프로필 관련 Go 호출. 화면은 바인딩 대신 이 파일을 쓴다.
import { ProfileService } from '../../bindings/github.com/TogetherBros/azit/desktop/internal/bridge'
import type { Profile } from '../components/ProfileCard'

// GET /v1/me 응답 (proto.MeResponse 와 같은 모양)
export type Me = {
  account_id: number
  name: string | null // null 이면 아직 프로필을 안 만듦
  nickname: string
  bio: string
  color: string
  avatar_url: string // 24시간짜리 보기 주소, 사진 없으면 ""
  banner_url: string
  needs_profile: boolean
}

export type ImageKind = 'avatar' | 'banner'

export async function getMe(): Promise<Me> {
  return (await ProfileService.Me()) as unknown as Me
}

// 고칠 수 있는 글자·색 (PATCH /v1/me 요청과 같은 모양). 네 칸을 다 보낸다.
export type ProfileFields = { name: string; nickname: string; bio: string; color: string }

// 프로필을 저장하고 저장된 내 정보를 받는다. 이름이 겹치면 "이미 누가 쓰고 있는 이름이에요" 오류.
export async function updateMe(f: ProfileFields): Promise<Me> {
  return (await ProfileService.UpdateMe(f)) as unknown as Me
}

export type NameCheck = { available: boolean; message?: string }

// 그 이름을 쓸 수 있는지. 지금 내 이름이면 available.
export async function checkName(name: string): Promise<NameCheck> {
  return (await ProfileService.CheckName(name)) as NameCheck
}

// 이름 규칙 (서버 account.ValidateName 과 같다). 서버에 묻기 전에 먼저 걸러 낸다.
export const NAME_RULE = /^[a-z0-9_]{5,20}$/

// 사진 하나를 올리고 내 프사·배너로 저장한다. Go 가 서버 → R2 → 서버 순서로 처리한다.
export async function uploadImage(kind: ImageKind, contentType: string, base64: string): Promise<void> {
  await ProfileService.UploadImage(kind, contentType, base64)
}

// 서버 응답을 출입증 카드가 쓰는 모양으로 바꾼다.
export function toProfile(me: Me): Profile {
  return {
    nickname: me.nickname,
    name: me.name ?? '',
    color: me.color,
    bio: me.bio || undefined,
    avatarSrc: me.avatar_url || undefined,
    bannerSrc: me.banner_url || undefined,
  }
}

export function profileErrorText(e: unknown): string {
  const msg = e instanceof Error ? e.message : String(e)
  if (/refused|dial tcp|connect/i.test(msg)) return '서버에 연결하지 못했어요. 서버가 켜져 있는지 봐 주세요'
  if (/[가-힣]/.test(msg)) return msg
  return '잠시 후 다시 해 주세요'
}
