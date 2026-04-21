package application

import "context"

type NewFollowerNotifier interface {
	Execute(ctx context.Context, input interface{}) error
}

var newFollowerNotifier NewFollowerNotifier

func SetNewFollowerNotifier(n NewFollowerNotifier) {
	newFollowerNotifier = n
}

func GetNewFollowerNotifier() NewFollowerNotifier {
	return newFollowerNotifier
}
