package persist

import (
	"sync"

	"simple-golang-crawler/engine"

	log "github.com/sirupsen/logrus"
)

func VideoItemCleaner(wgOutside *sync.WaitGroup) (chan *engine.Item, error) {
	out := make(chan *engine.Item)
	go func() {
		defer wgOutside.Done()
		itemCount := 0
		for item := range out {
			log.Printf("Item Saver:got item "+
				"#%d: %v", itemCount, item)
			itemCount++
		}
	}()
	return out, nil
}
