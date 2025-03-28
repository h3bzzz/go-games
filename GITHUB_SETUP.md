# GitHub Repository Setup

## 1. Create a New Repository on GitHub

1. Visit https://github.com/new
2. Set Repository name to "go-chess"
3. Add a description: "A chess game written in Go with Fyne UI and AI opponents"
4. Make sure the repository is set to "Public"
5. Do not initialize with README, .gitignore, or license (we've created these locally)
6. Click "Create repository"

## 2. Push Your Local Repository to GitHub

After creating the repository on GitHub, use the following commands to push your local repository:

```bash
# Add the remote repository
git remote add origin https://github.com/h3bzzz/go-chess.git

# Push your local repository to GitHub
git push -u origin main
```

## 3. Add Screenshots

For the README screenshots to display properly, you'll need to:

1. Take screenshots of your chess game running with different themes
2. Resize them to a reasonable size (e.g., 800x600 pixels)
3. Save them in the assets/screenshots directory with the following names:
   - game.png (main game screenshot)
   - classic.png (classic theme)
   - green.png (green theme)
   - pink.png (pink theme)
4. Push these changes to GitHub:

```bash
git add assets/screenshots/
git commit -m "Add game screenshots"
git push
```

## 4. Verify the README

After pushing your repository to GitHub:

1. Visit your repository at https://github.com/h3bzzz/go-chess
2. Check that the README is properly displayed with images
3. Make any necessary adjustments to improve the presentation

## 5. Add GitHub Topics (optional)

To make your repository more discoverable:

1. Click on the gear icon next to "About" on your repository page
2. Add relevant topics such as: go, golang, chess, game, fyne, ai, gui
3. Click "Save changes" 