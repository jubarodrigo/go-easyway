package cmd

func StartApp() {
	star := NewStarship()
	buildDirector := NewStarshipBuilder(star)

	buildDirector.BuildStarship()
}
