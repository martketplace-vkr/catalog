package vendor

import "testing"

func TestValidateCryptoPricing(t *testing.T) {
	tests := []struct {
		name          string
		acceptsCrypto bool
		mode          string
		priceUSDT     string
		wantErr       bool
	}{
		{name: "disabled", mode: "disabled"},
		{name: "fixed price", acceptsCrypto: true, mode: "fixed_usdt", priceUSDT: "12.34567890"},
		{name: "rate based", acceptsCrypto: true, mode: "rub_rate"},
		{name: "fixed missing price", acceptsCrypto: true, mode: "fixed_usdt", wantErr: true},
		{name: "fixed zero price", acceptsCrypto: true, mode: "fixed_usdt", priceUSDT: "0", wantErr: true},
		{name: "disabled with price", mode: "disabled", priceUSDT: "10", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCryptoPricing(tt.acceptsCrypto, tt.mode, tt.priceUSDT)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateCryptoPricing() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
