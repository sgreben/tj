package main

type tokenBuffer []tokenStream

func (b *tokenBuffer) flush(token tokenStream) {
	for i, oldToken := range *b {
		oldToken.Token().copyDeltasFrom(token.Token())
		print(oldToken)
		(*b)[i] = nil
	}
	*b = (*b)[:0]
}
