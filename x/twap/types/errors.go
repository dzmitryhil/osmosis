package types

import (
	"errors"
	"fmt"
	time "time"
)

var (
	ErrEmptyRoute           = errors.New("route cannot be empty")
	ErrMismatchedQuoteAsset = errors.New("last pool's quote asset must match the quote asset provided")
)

type EndTimeInFutureError struct {
	EndTime   time.Time
	BlockTime time.Time
}

func (e EndTimeInFutureError) Error() string {
	return fmt.Sprintf("called GetArithmeticTwap with an end time in the future."+
		" (end time %s, current time %s)", e.EndTime, e.BlockTime)
}

type StartTimeAfterEndTimeError struct {
	StartTime time.Time
	EndTime   time.Time
}

func (e StartTimeAfterEndTimeError) Error() string {
	return fmt.Sprintf("called GetArithmeticTwap with a start time that is after the end time."+
		" (start time %s, end time %s)", e.StartTime, e.EndTime)
}

type KeySeparatorLengthError struct {
	ExpectedLength int
	ActualLength   int
}

func (e KeySeparatorLengthError) Error() string {
	return fmt.Sprintf("key separator is an incorrect length."+
		" (expected length %d, actual length %d)", e.ExpectedLength, e.ActualLength)
}

type UnexpectedSeparatorError struct {
	ExpectedSeparator string
	ActualSeparator   string
}

func (e UnexpectedSeparatorError) Error() string {
	return fmt.Sprintf("separator is incorrectly formatted."+
		" (expected separator %s, actual separator %v)", e.ExpectedSeparator, e.ActualSeparator)
}

type TimeStringKeyFormatError struct {
	Key string
	Err error
}

func (e TimeStringKeyFormatError) Error() string {
	return fmt.Sprintf("incorrectly formatted time string in key %s : %v", e.Key, e.Err)
}

type InvalidRecordCountError struct {
	Actual   int
	Expected int
}

func (e InvalidRecordCountError) Error() string {
	return fmt.Sprintf("The number of records do not match, expected: %d\n got: %d", e.Expected, e.Actual)
}

type InvalidUpdateRecordError struct {
	RecordBlockHeight int64
	RecordTime        time.Time
	ActualBlockHeight int64
	ActualTime        time.Time
}

func (e InvalidUpdateRecordError) Error() string {
	return fmt.Sprintf("failed to update the record, the context time must be greater than record time; record: block %d at %s, actual: block %d at %s", e.RecordBlockHeight, e.RecordTime, e.ActualBlockHeight, e.ActualTime)
}
