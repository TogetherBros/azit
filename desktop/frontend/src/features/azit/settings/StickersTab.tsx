import { useEffect, useRef, useState } from 'react'
import type { Sticker } from '../../../api/types'
import { Icon } from '../../../components/Icon'

type Props = {
  stickers: Sticker[]
  uploading?: boolean
  error?: string
  canDelete?: (id: string) => boolean // 올린 사람이나 관리자만
  onUpload?: (file: File, name: string) => Promise<boolean> | void // 성공하면 true
  onDelete?: (id: string) => void
}

// 짤: 아지트가 같이 쓰는 짤 모음. 올리기 · 빼기. (맞장구 목록은 "맞장구" 탭)
export function StickersTab({ stickers, uploading, error, canDelete, onUpload, onDelete }: Props) {
  const fileRef = useRef<HTMLInputElement>(null)
  const [file, setFile] = useState<File>()
  const [name, setName] = useState('')
  const [src, setSrc] = useState<string>()

  useEffect(() => {
    if (!file) { setSrc(undefined); return }
    const url = URL.createObjectURL(file)
    setSrc(url)
    return () => URL.revokeObjectURL(url)
  }, [file])

  const pick = (f?: File) => {
    if (!f) return
    setFile(f)
    setName(f.name.replace(/\.[^.]+$/, '').slice(0, 20))
  }

  const submit = async () => {
    if (!file || !name.trim() || !onUpload) return
    const ok = await onUpload(file, name.trim())
    if (ok) { setFile(undefined); setName('') }
  }

  return (
    <>
      <h1 className="page-title sm">짤</h1>
      <span className="hint">아지트 멤버 모두가 같이 써요. 맞장구로 달 수도 있어요 (맞장구 목록에 없어도 돼요).</span>

      <section className="card col" style={{ gap: 12 }}>
        <div className="row">
          <span className="grow b">짤 {stickers.length}</span>
          <button type="button" className="btn p" disabled={uploading} onClick={() => fileRef.current?.click()}>+ 짤 올리기</button>
          <input ref={fileRef} type="file" accept="image/*" hidden onChange={(e) => { pick(e.target.files?.[0]); e.target.value = '' }} />
        </div>

        {file && (
          <form className="row card2 meme-new" noValidate onSubmit={(e) => { e.preventDefault(); void submit() }}>
            {src && <img src={src} alt="" />}
            <div className="grow col" style={{ gap: 6 }}>
              <label htmlFor="meme-name" className="lbl">짤 이름</label>
              <input id="meme-name" className="inp" maxLength={20} value={name} onChange={(e) => setName(e.target.value)} autoFocus />
            </div>
            <button type="submit" className="btn p" disabled={!name.trim() || uploading}>{uploading ? '올리는 중…' : '올리기'}</button>
            <button type="button" className="btn sec" disabled={uploading} onClick={() => setFile(undefined)}>취소</button>
          </form>
        )}
        {error && <span role="alert" className="ferr">{error}</span>}
        <span className="hint">긴 변을 512px 로 줄여서 저장해요.</span>

        {stickers.length
          ? (
            <div className="st-cards">
              {stickers.map((st) => (
                <div key={st.id} className="card2 st-card">
                  <img src={st.src} alt={st.name} />
                  <span className="b st-card-name">{st.name}</span>
                  <span className="mu" style={{ fontSize: 11 }}>{st.by ? `${st.by} · ` : ''}{st.ago}</span>
                  {(canDelete?.(st.id) ?? true) && (
                    <button type="button" className="btn s dg" aria-label={`짤 빼기 ${st.name}`} onClick={onDelete && (() => onDelete(st.id))}>
                      <Icon name="x" size={14} /> 빼기
                    </button>
                  )}
                </div>
              ))}
            </div>
          )
          : <div className="card2 hint" style={{ textAlign: 'center', padding: 24 }}>아직 짤이 없어요</div>}
      </section>
    </>
  )
}
