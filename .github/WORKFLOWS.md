# GitHub Actions CI/CD

이 프로젝트는 GitHub Actions를 사용하여 자동화된 CI/CD 파이프라인을 제공합니다.

## 워크플로우

### 1. CI (Continuous Integration)

**트리거:** `main` 브랜치에 push 또는 Pull Request 생성 시

**작업:**
- **Test**: 모든 테스트 실행 및 커버리지 리포트 생성
  - Go 1.25 사용
  - Race detector 활성화
  - Coverage 리포트를 Codecov에 업로드
  - Coverage HTML 리포트를 artifact로 저장 (30일 보관)
  
- **Build**: 5개 플랫폼용 바이너리 빌드
  - linux-amd64
  - linux-arm64
  - darwin-amd64 (Intel Mac)
  - darwin-arm64 (Apple Silicon)
  - windows-x86_64
  - 버전 정보 주입 (dev-{commit})
  - 빌드된 바이너리를 artifact로 저장 (7일 보관)
  
- **Lint**: 코드 품질 검사
  - golangci-lint 실행
  - 설정: `.golangci.yml`

### 2. Release (Manual Dispatch)

**트리거:** GitHub UI에서 수동 실행 (Actions → Release → Run workflow)

**입력 파라미터:**
- `version`: 릴리스 버전 (예: v1.0.0)
- `prerelease`: Pre-release 여부 (기본값: false)

**작업:**
- **Create Release**:
  - 버전 형식 검증 (v1.2.3 또는 v1.2.3-beta.1)
  - Git 태그 생성
  - 자동 changelog 생성
  - GitHub Release 생성
  
- **Build and Upload**:
  - 5개 플랫폼용 바이너리 빌드 및 압축
  - Linux/macOS: tar.gz
  - Windows: zip
  - SHA256 체크섬 생성
  - Release에 asset 업로드
  
- **Notify**: 릴리스 완료 알림

## 사용 방법

### CI 실행

자동으로 실행됩니다:
```bash
# main 브랜치에 push
git push origin main

# Pull Request 생성
gh pr create --base main
```

### Release 생성

1. GitHub 웹사이트로 이동
2. Actions 탭 클릭
3. "Release" 워크플로우 선택
4. "Run workflow" 버튼 클릭
5. 버전 입력 (예: v1.0.0)
6. Pre-release 여부 선택
7. "Run workflow" 실행

또는 GitHub CLI 사용:
```bash
# 정식 릴리스
gh workflow run release.yml -f version=v1.0.0 -f prerelease=false

# Pre-release
gh workflow run release.yml -f version=v1.0.0-beta.1 -f prerelease=true
```

### Coverage 리포트 확인

1. **GitHub Actions Artifacts**:
   - Actions 탭 → CI 워크플로우 → 실행 선택
   - Artifacts 섹션에서 `coverage-report` 다운로드
   - `coverage.html` 파일을 브라우저로 열기

2. **Codecov (설정 필요)**:
   - https://codecov.io/gh/dh-kam/go-cert-provider
   - Repository에 Codecov 토큰 추가:
     - Settings → Secrets → New repository secret
     - Name: `CODECOV_TOKEN`
     - Value: Codecov에서 생성된 토큰

### Lint 로컬 실행

```bash
# golangci-lint 설치
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Lint 실행
golangci-lint run

# 자동 수정 가능한 이슈 수정
golangci-lint run --fix
```

## 배지 (Badges)

README.md에 추가된 배지:

- **CI Status**: CI 워크플로우 상태
- **Code Coverage**: Codecov 커버리지 비율
- **Go Report Card**: Go 코드 품질 점수
- **License**: 라이선스 정보

## 파일 구조

```
.github/
├── workflows/
│   ├── ci.yml          # CI 워크플로우
│   └── release.yml     # Release 워크플로우
.golangci.yml           # Linter 설정
codecov.yml             # Codecov 설정
```

## 주의사항

### Release 버전 형식

- ✅ `v1.0.0`
- ✅ `v1.2.3-beta.1`
- ✅ `v2.0.0-rc.1`
- ❌ `1.0.0` (v 접두사 없음)
- ❌ `v1.0` (패치 버전 없음)
- ❌ `latest` (의미론적 버전 아님)

### Secrets 설정

선택적으로 설정 가능:
- `CODECOV_TOKEN`: Codecov 통합 (없어도 CI는 실패하지 않음)

### 빌드 산출물

- **CI Artifacts**: 7일 보관
- **Coverage Reports**: 30일 보관
- **Release Assets**: 영구 보관

## 트러블슈팅

### CI 실패 시

1. **Test 실패**:
   ```bash
   # 로컬에서 테스트 실행
   go test -v -race ./...
   ```

2. **Build 실패**:
   ```bash
   # 로컬에서 빌드 테스트
   GOOS=linux GOARCH=amd64 go build .
   ```

3. **Lint 실패**:
   ```bash
   # 로컬에서 lint 실행
   golangci-lint run
   ```

### Release 실패 시

- 버전 태그가 이미 존재하는지 확인
- 버전 형식이 올바른지 확인 (v1.2.3)
- Git history가 있는지 확인 (changelog 생성용)

## 개선 사항

향후 추가 가능한 기능:
- [ ] Docker 이미지 빌드 및 푸시
- [ ] 자동 릴리스 노트 생성 (conventional commits)
- [ ] 보안 스캔 (Trivy, Snyk)
- [ ] 성능 벤치마크 추적
- [ ] 자동 버전 관리 (semantic-release)
