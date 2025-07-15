package ginserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

func FetchMetadata(filename string) (*ChunkMetadata, error) {
	url := fmt.Sprintf("http://torrent-server/metadata/%s", filename)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-200 response: %d", resp.StatusCode)
	}

	var meta ChunkMetadata
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return nil, fmt.Errorf("error decoding metadata: %w", err)
	}

	return &meta, nil
}

func DownloadChunk(filename string, index int) error {
	url := fmt.Sprintf("http://torrent-server/chunk/%s/%d", filename, index)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	os.MkdirAll("chunks", os.ModePerm)
	outPath := filepath.Join("chunks", fmt.Sprintf("%s.chunk.%d", filename, index))
	outFile, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	return err
}

func DownloadMyChunks(filename string, leecherIndex, totalLeechers int) error {
	meta, err := FetchMetadata(filename)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for i := 0; i < meta.NumChunks; i++ {
		if i%totalLeechers == leecherIndex {
			wg.Add(1)
			go func(chunkIndex int) {
				defer wg.Done()
				if err := DownloadChunk(filename, chunkIndex); err != nil {
					fmt.Printf("Failed to download chunk %d: %v\n", chunkIndex, err)
				} else {
					fmt.Printf("Downloaded chunk %d\n", chunkIndex)
				}
			}(i)
		}
	}
	wg.Wait()
	return nil
}
