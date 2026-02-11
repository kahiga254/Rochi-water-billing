package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"
)

// GenerateBillNumber generates a unique bill number
func GenerateBillNumber() string {
	now := time.Now()
	timestamp := now.Format("20060102")

	// Add random 4 digits for uniqueness
	randomNum, _ := rand.Int(rand.Reader, big.NewInt(9000))
	randomNum = randomNum.Add(randomNum, big.NewInt(1000))

	return fmt.Sprintf("BILL-%s-%s", timestamp, randomNum.String())
}

// GenerateReceiptNumber generates a unique receipt number
func GenerateReceiptNumber() string {
	now := time.Now()
	timestamp := now.Format("20060102")

	randomNum, _ := rand.Int(rand.Reader, big.NewInt(9000))
	randomNum = randomNum.Add(randomNum, big.NewInt(1000))

	return fmt.Sprintf("RCPT-%s-%s", timestamp, randomNum.String())
}

// FormatPhoneNumber formats phone number to E.164 format
func FormatPhoneNumber(phone string) string {
	// Remove any non-digit characters
	phone = strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, phone)

	// If starts with 0, replace with country code
	if strings.HasPrefix(phone, "0") {
		phone = "254" + phone[1:]
	}

	// If doesn't start with +, add it
	if !strings.HasPrefix(phone, "+") {
		phone = "+" + phone
	}

	return phone
}

// ValidateMeterNumber validates meter number format
func ValidateMeterNumber(meterNumber string) bool {
	// Basic validation - can be extended based on your meter number format
	if len(meterNumber) < 3 || len(meterNumber) > 20 {
		return false
	}

	// Should contain only alphanumeric characters
	for _, char := range meterNumber {
		if !((char >= '0' && char <= '9') ||
			(char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z')) {
			return false
		}
	}

	return true
}

// CalculateConsumption calculates water consumption
func CalculateConsumption(previous, current float64) (float64, error) {
	if current < previous {
		return 0, fmt.Errorf("current reading (%.2f) cannot be less than previous reading (%.2f)", current, previous)
	}
	return current - previous, nil
}

// CalculateAmount calculates total amount including all charges
func CalculateAmount(consumption, rate, fixedCharge, arrears, penalty, discount float64) float64 {
	waterCharge := consumption * rate
	total := waterCharge + fixedCharge + arrears + penalty - discount

	// Ensure minimum amount is 0
	if total < 0 {
		return 0
	}
	return total
}

// GetBillingPeriod returns the billing period string
func GetBillingPeriod(date time.Time) string {
	return date.Format("January 2006")
}

// GetMonthYear returns month and year from date
func GetMonthYear(date time.Time) (string, int) {
	return date.Format("2006-01"), date.Year()
}

// RoundToTwoDecimal rounds float to 2 decimal places
func RoundToTwoDecimal(value float64) float64 {
	rounded, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", value), 64)
	return rounded
}

// ParseDateString parses date string in various formats
func ParseDateString(dateStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02",
		"02/01/2006",
		"01/02/2006",
		time.RFC3339,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}
