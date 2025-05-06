package co

import "iter"

type Iterator[T, K any] iter.Seq2[K, IterResultItem[T]]

func (ir Iterator[T, K]) CollectAll() (result PartialResult[T, K]) {
	for idx, item := range ir {
		if item.Err != nil {
			result.addError(idx, item.Err)
		} else {
			result.addResult(idx, item.Result)
		}
	}

	return
}

func (ir Iterator[T, K]) CollectAllResultsOrFirstError() (result PartialResult[T, K]) {
	for key, item := range ir {
		if err := item.Err; err != nil {
			result.addError(key, err)
			result.Results = nil
			return
		}

		result.addResult(key, item.Result)
	}

	return
}

func (ir Iterator[T, K]) CollectFirstResult() (result PartialResult[T, K]) {
	for key, item := range ir {
		if err := item.Err; err != nil {
			result.addError(key, err)
		} else {
			result.addResult(key, item.Result)
			result.Errors = nil
			return
		}
	}

	return
}
