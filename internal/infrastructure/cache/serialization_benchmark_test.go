package cache

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/dowglassantana/product-redis-api/internal/domain/entity"
	"github.com/vmihailenco/msgpack/v5"
)

// Produto de teste com dados realistas
func createTestProduct() *entity.Product {
	return &entity.Product{
		ID:              "01HN8Z9QXXXXXXXXXXXXXXXXXXXX",
		Name:            "iPhone 15 Pro Max 256GB Titanium Natural",
		ReferenceNumber: "APPLE-IP15PM-256-TIT-NAT",
		Category:        "Smartphones",
		Description:     "O iPhone 15 Pro Max possui um design em titânio, chip A17 Pro, sistema de câmera profissional com zoom óptico de 5x, e recursos avançados de segurança. Tela Super Retina XDR de 6,7 polegadas com ProMotion.",
		SKU:             "SKU-APPLE-IP15PM-256-TIT-NAT-2024",
		Brand:           "Apple",
		Stock:           150,
		Images: []string{
			"https://example.com/images/iphone15promax-front.jpg",
			"https://example.com/images/iphone15promax-back.jpg",
			"https://example.com/images/iphone15promax-side.jpg",
			"https://example.com/images/iphone15promax-camera.jpg",
		},
		Specifications: map[string]interface{}{
			"storage":      "256GB",
			"color":        "Titanium Natural",
			"chip":         "A17 Pro",
			"display":      "6.7-inch Super Retina XDR",
			"camera":       "48MP Main + 12MP Ultra Wide + 12MP Telephoto",
			"battery":      "4422mAh",
			"water_resist": "IP68",
			"5g":           true,
			"weight":       "221g",
			"dimensions":   "159.9 x 76.7 x 8.25 mm",
		},
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Benchmark de serialização JSON (Marshal)
func BenchmarkJSONMarshal(b *testing.B) {
	product := createTestProduct()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(product)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark de deserialização JSON (Unmarshal)
func BenchmarkJSONUnmarshal(b *testing.B) {
	product := createTestProduct()
	data, _ := json.Marshal(product)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var p entity.Product
		err := json.Unmarshal(data, &p)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark de ciclo completo JSON (Marshal + Unmarshal)
func BenchmarkJSONRoundTrip(b *testing.B) {
	product := createTestProduct()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		data, err := json.Marshal(product)
		if err != nil {
			b.Fatal(err)
		}

		var p entity.Product
		err = json.Unmarshal(data, &p)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark de deserialização de múltiplos produtos (simula GetMultiple)
func BenchmarkJSONUnmarshalMultiple(b *testing.B) {
	// Simula 50 produtos (tamanho típico de uma listagem)
	products := make([][]byte, 50)
	for i := 0; i < 50; i++ {
		product := createTestProduct()
		product.ID = product.ID[:20] + string(rune('A'+i%26)) + string(rune('A'+i/26))
		data, _ := json.Marshal(product)
		products[i] = data
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result := make([]*entity.Product, 0, 50)
		for _, data := range products {
			var p entity.Product
			err := json.Unmarshal(data, &p)
			if err != nil {
				b.Fatal(err)
			}
			result = append(result, &p)
		}
	}
}

// Mede o tamanho do payload JSON
func TestJSONPayloadSize(t *testing.T) {
	product := createTestProduct()
	data, _ := json.Marshal(product)
	t.Logf("JSON payload size: %d bytes", len(data))
}

// ==================== MSGPACK BENCHMARKS ====================

// Benchmark de serialização Msgpack (Marshal)
func BenchmarkMsgpackMarshal(b *testing.B) {
	product := createTestProduct()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := msgpack.Marshal(product)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark de deserialização Msgpack (Unmarshal)
func BenchmarkMsgpackUnmarshal(b *testing.B) {
	product := createTestProduct()
	data, _ := msgpack.Marshal(product)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var p entity.Product
		err := msgpack.Unmarshal(data, &p)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark de ciclo completo Msgpack (Marshal + Unmarshal)
func BenchmarkMsgpackRoundTrip(b *testing.B) {
	product := createTestProduct()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		data, err := msgpack.Marshal(product)
		if err != nil {
			b.Fatal(err)
		}

		var p entity.Product
		err = msgpack.Unmarshal(data, &p)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark de deserialização de múltiplos produtos com Msgpack
func BenchmarkMsgpackUnmarshalMultiple(b *testing.B) {
	products := make([][]byte, 50)
	for i := 0; i < 50; i++ {
		product := createTestProduct()
		product.ID = product.ID[:20] + string(rune('A'+i%26)) + string(rune('A'+i/26))
		data, _ := msgpack.Marshal(product)
		products[i] = data
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result := make([]*entity.Product, 0, 50)
		for _, data := range products {
			var p entity.Product
			err := msgpack.Unmarshal(data, &p)
			if err != nil {
				b.Fatal(err)
			}
			result = append(result, &p)
		}
	}
}

// Mede o tamanho do payload Msgpack
func TestMsgpackPayloadSize(t *testing.T) {
	product := createTestProduct()
	data, _ := msgpack.Marshal(product)
	t.Logf("Msgpack payload size: %d bytes", len(data))
}

// Compara tamanho dos payloads
func TestPayloadSizeComparison(t *testing.T) {
	product := createTestProduct()

	jsonData, _ := json.Marshal(product)
	msgpackData, _ := msgpack.Marshal(product)

	reduction := float64(len(jsonData)-len(msgpackData)) / float64(len(jsonData)) * 100

	t.Logf("JSON payload size: %d bytes", len(jsonData))
	t.Logf("Msgpack payload size: %d bytes", len(msgpackData))
	t.Logf("Size reduction: %.2f%%", reduction)
}
