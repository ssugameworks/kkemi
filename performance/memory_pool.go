package performance

import (
	"discord-bot/constants"
	"discord-bot/models"
	"strings"
	"sync"
)

var (
	// ScoreDataSlicePool 점수 데이터 슬라이스 풀
	ScoreDataSlicePool = sync.Pool{
		New: func() interface{} {
			// 기본 용량으로 슬라이스 생성
			slice := make([]models.ScoreData, 0, 50)
			return &slice
		},
	}

	// ScoreDataChanPool 점수 데이터 채널 풀
	ScoreDataChanPool = sync.Pool{
		New: func() interface{} {
			// 기본 버퍼 크기로 채널 생성
			ch := make(chan models.ScoreData, 100)
			return ch
		},
	}

	// SemaphoreChanPool 세마포어 채널 풀
	SemaphoreChanPool = sync.Pool{
		New: func() interface{} {
			// 최대 동시성 크기로 채널 생성
			ch := make(chan struct{}, 20)
			return ch
		},
	}

	// StringBuilderPool 문자열 빌더 풀 (로그 및 메시지 생성용)
	StringBuilderPool = sync.Pool{
		New: func() interface{} {
			return &strings.Builder{}
		},
	}
)

// GetScoreDataSlice 재사용 가능한 점수 데이터 슬라이스를 가져옵니다
func GetScoreDataSlice() *[]models.ScoreData {
	slice := ScoreDataSlicePool.Get().(*[]models.ScoreData)
	// 슬라이스 초기화 (길이는 0으로, 용량은 유지)
	*slice = (*slice)[:0]
	return slice
}

// PutScoreDataSlice 점수 데이터 슬라이스를 풀에 반환합니다
func PutScoreDataSlice(slice *[]models.ScoreData) {
	// 메모리 누수 방지를 위해 큰 슬라이스는 풀에 반환하지 않음
	if cap(*slice) <= constants.MaxPoolSliceCapacity {
		ScoreDataSlicePool.Put(slice)
	}
}

// GetScoreDataChannel 재사용 가능한 점수 데이터 채널을 가져옵니다
func GetScoreDataChannel(bufferSize int) chan models.ScoreData {
	if bufferSize <= constants.MaxPoolChannelCapacity {
		// 풀에서 재사용 가능한 채널 가져오기
		ch := ScoreDataChanPool.Get().(chan models.ScoreData)
		// 채널이 비어있는지 확인하고 비우기
		for {
			select {
			case <-ch:
				// 채널에서 남은 데이터 제거
			default:
				return ch
			}
		}
	}
	// 큰 버퍼가 필요하면 새로 생성
	return make(chan models.ScoreData, bufferSize)
}

// PutScoreDataChannel 점수 데이터 채널을 풀에 반환합니다
func PutScoreDataChannel(ch chan models.ScoreData) {
	if cap(ch) <= constants.MaxPoolChannelCapacity {
		// 채널이 닫혔는지 확인
		select {
		case _, ok := <-ch:
			if !ok {
				// 채널이 닫혔으므로 풀에 반환하지 않음
				return
			}
			// 받은 데이터는 버리고 계속해서 채널을 비움
		default:
			// 채널이 비어있음
		}
		
		// 채널을 완전히 비운 후 풀에 반환
		for {
			select {
			case _, ok := <-ch:
				if !ok {
					// 채널이 닫혔으므로 풀에 반환하지 않음
					return
				}
				// 채널에서 모든 데이터 제거
			default:
				ScoreDataChanPool.Put(ch)
				return
			}
		}
	}
}

// GetSemaphoreChannel 재사용 가능한 세마포어 채널을 가져옵니다
func GetSemaphoreChannel(size int) chan struct{} {
	if size <= constants.MaxPoolSemaphoreSize {
		ch := SemaphoreChanPool.Get().(chan struct{})
		// 채널이 비어있는지 확인하고 비우기
		for {
			select {
			case <-ch:
				// 채널에서 남은 토큰 제거
			default:
				return ch
			}
		}
	}
	// 큰 세마포어가 필요하면 새로 생성
	return make(chan struct{}, size)
}

// PutSemaphoreChannel 세마포어 채널을 풀에 반환합니다
func PutSemaphoreChannel(ch chan struct{}) {
	if cap(ch) <= constants.MaxPoolSemaphoreSize {
		// 채널을 비운 후 풀에 반환
		for {
			select {
			case <-ch:
				// 채널에서 모든 토큰 제거
			default:
				SemaphoreChanPool.Put(ch)
				return
			}
		}
	}
}

// GetStringBuilder 재사용 가능한 문자열 빌더를 가져옵니다
func GetStringBuilder() *strings.Builder {
	sb := StringBuilderPool.Get().(*strings.Builder)
	sb.Reset() // 내용 초기화
	return sb
}

// PutStringBuilder 문자열 빌더를 풀에 반환합니다
func PutStringBuilder(sb *strings.Builder) {
	// 너무 큰 빌더는 풀에 반환하지 않음 (메모리 누수 방지)
	if sb.Cap() <= constants.MaxStringBuilderSize {
		StringBuilderPool.Put(sb)
	}
}

// PoolStats 메모리 풀 통계 정보
type PoolStats struct {
	ScoreDataSlicePoolSize int
	ScoreDataChanPoolSize  int
	SemaphoreChanPoolSize  int
	StringBuilderPoolSize  int
}

// GetPoolStats 현재 메모리 풀 통계를 반환합니다 (대략적인 값)
func GetPoolStats() PoolStats {
	// sync.Pool은 내부 통계를 제공하지 않으므로 대략적인 추정치 반환
	return PoolStats{
		ScoreDataSlicePoolSize: 0, // 정확한 값을 얻기 어려움
		ScoreDataChanPoolSize:  0, // 정확한 값을 얻기 어려움
		SemaphoreChanPoolSize:  0, // 정확한 값을 얻기 어려움
		StringBuilderPoolSize:  0, // 정확한 값을 얻기 어려움
	}
}