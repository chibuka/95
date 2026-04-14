package banner

import "github.com/charmbracelet/lipgloss"

const art = `
________   ___  ________   _______   ________ ___  ___      ___ _______
|\   ___  \|\  \|\   ___  \|\  ___ \ |\  _____\\  \|\  \    /  /|\  ___ \
\ \  \\ \  \ \  \ \  \\ \  \ \   __/|\ \  \__/\ \  \ \  \  /  / | \   __/|
\ \  \\ \  \ \  \ \  \\ \  \ \  \_|/_\ \   __\\ \  \ \  \/  / / \ \  \_|/__
 \ \  \\ \  \ \  \ \  \\ \  \ \  \_|\ \ \  \_| \ \  \ \    / /   \ \  \_|\ \
  \ \__\\ \__\ \__\ \__\\ \__\ \_______\ \__\   \ \__\ \__/ /     \ \_______\
   \|__| \|__|\|__|\|__| \|__|\|_______|\|__|    \|__|\|__|/       \|_______|`

var artStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("242"))

// Render returns the styled ASCII art string.
func Render() string {
	return artStyle.Render(art)
}
