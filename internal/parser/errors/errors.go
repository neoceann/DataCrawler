package parserErrors

import "errors"

var ErrEmptyLinks = errors.New("list of links are empty")

var ErrEmptyDataForProduct = errors.New("empty data for this product")
