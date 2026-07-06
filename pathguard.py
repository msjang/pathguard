#!/usr/bin/env python3
"""경로/파일명 길이 가드 (읽기전용).
유니코드 NFD(조합형) 최악치 바이트로 환산해 NAME_MAX/PATH_MAX 초과·임박 파일을 찾는다.
NFD 분해로 바이트가 늘어나는 건 한글이 가장 극단적이지만(3배) 베트남어·일본어 탁점가나·
라틴 악센트 등 결합 문자 전반에서 일어난다. b_nfd()는 언어 무관하게 동작한다.
동기화 대상 파일시스템(예: NAS의 btrfs/ext4)의 한계: NAME_MAX 255, PATH_MAX 4096 (바이트).
설정은 환경변수로 덮어쓸 수 있다 (PATHGUARD_* 참고).
"""
import os, sys, unicodedata, json

# 동기화되지 않거나 도구/서버가 생성하는 노이즈 디렉터리·파일 (기본 제외).
# 예: @eaDir=시놀로지 캐시(서버 생성), #recycle=시놀 휴지통, .git=보통 동기화 제외.
DEFAULT_EXCLUDE = {
    ".git", "node_modules",
    "@eaDir", "#recycle", "#snapshot",              # Synology
    ".DS_Store", ".Trashes", ".Spotlight-V100", ".fseventsd",  # macOS
    "$RECYCLE.BIN", "System Volume Information",     # Windows
}

ROOT          = os.path.expanduser(os.environ.get("PATHGUARD_ROOT", "~/Documents"))
REMOTE_PREFIX = os.environ.get("PATHGUARD_REMOTE_PREFIX", "/volume1/homes/johndoe/MyDocuments")  # 원격(NAS/클라우드) 쪽 절대경로 루트
NAME_MAX      = int(os.environ.get("PATHGUARD_NAME_MAX", "255"))    # 바이트, 경로 구성요소(파일/폴더명) 하나당
PATH_MAX      = int(os.environ.get("PATHGUARD_PATH_MAX", "4096"))   # 바이트, 전체 경로
WARN          = float(os.environ.get("PATHGUARD_WARN", "0.80"))     # 한계의 80%부터 경고
# PATHGUARD_EXCLUDE=쉼표구분 이름목록 (설정 시 기본값을 대체). 빈 문자열이면 제외 없음.
_env_excl     = os.environ.get("PATHGUARD_EXCLUDE")
EXCLUDE       = ({e.strip() for e in _env_excl.split(",") if e.strip()}
                 if _env_excl is not None else set(DEFAULT_EXCLUDE))

def b_nfc(s): return len(unicodedata.normalize('NFC', s).encode('utf-8'))
def b_nfd(s): return len(unicodedata.normalize('NFD', s).encode('utf-8'))
def form(s):
    if s == unicodedata.normalize('NFC', s): return 'NFC'
    if s == unicodedata.normalize('NFD', s): return 'NFD'
    return 'mixed'

def scan(root=ROOT):
    name_over, name_warn, path_over, path_warn = [], [], [], []
    total = 0
    for dp, dns, fns in os.walk(root):
        dns[:] = [d for d in dns if d not in EXCLUDE]   # 제외 폴더로는 내려가지 않음
        for name in list(dns) + list(fns):
            if name in EXCLUDE: continue                # 제외 파일 건너뜀
            total += 1
            full = os.path.join(dp, name)
            rel  = os.path.relpath(full, root)
            nfd_name = b_nfd(name)                    # 구성요소 최악치
            remote = REMOTE_PREFIX + "/" + rel
            nfd_path = b_nfd(remote)                  # 전체경로 최악치(원격)
            rec = {
                'rel': rel, 'form': form(name),
                'name_cur': len(name.encode('utf-8')), 'name_nfc': b_nfc(name), 'name_nfd': nfd_name,
                'path_nfd': nfd_path,
            }
            if nfd_name > NAME_MAX:      name_over.append(rec)
            elif nfd_name >= NAME_MAX*WARN: name_warn.append(rec)
            if nfd_path > PATH_MAX:      path_over.append(rec)
            elif nfd_path >= PATH_MAX*WARN: path_warn.append(rec)
    return total, name_over, name_warn, path_over, path_warn

def main():
    total, no, nw, po, pw = scan()
    no.sort(key=lambda r:-r['name_nfd']); nw.sort(key=lambda r:-r['name_nfd'])
    print(f"스캔: {total}개 항목 (한계 NAME_MAX={NAME_MAX}B, PATH_MAX={PATH_MAX}B, NFD 최악치 기준)")
    print(f"  파일/폴더명 초과(>{NAME_MAX}B): {len(no)}  | 경고({int(NAME_MAX*WARN)}~{NAME_MAX}B): {len(nw)}")
    print(f"  전체경로 초과(>{PATH_MAX}B): {len(po)}  | 경고: {len(pw)}")
    def show(rec):
        return (f"    NFD {rec['name_nfd']:>3}B (현재 {rec['name_cur']}B/{rec['form']}, NFC {rec['name_nfc']}B)"
                f"  {rec['rel']}")
    if no:
        print(f"\n■ NAME_MAX 초과 {len(no)}건:")
        for r in no[:40]: print(show(r))
    if nw:
        print(f"\n■ NAME_MAX 경고 {len(nw)}건 (상위 15):")
        for r in nw[:15]: print(show(r))
    if po:
        print(f"\n■ PATH_MAX 초과 {len(po)}건:")
        for r in po[:20]: print(f"    NFD {r['path_nfd']}B  {r['rel']}")
    # 요약 JSON (알림/스케줄용)
    summary = {'total': total, 'name_over': len(no), 'name_warn': len(nw),
               'path_over': len(po), 'path_warn': len(pw)}
    if len(sys.argv) > 1 and sys.argv[1] == '--json':
        print(json.dumps(summary, ensure_ascii=False))
    return summary

if __name__ == '__main__':
    main()
