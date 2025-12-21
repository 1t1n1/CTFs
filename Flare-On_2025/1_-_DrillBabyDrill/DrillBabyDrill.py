import pygame
import random

pygame.init()
pygame.font.init()
screen_width = 800
screen_height = 600
tile_size = 40
tiles_width = screen_width // tile_size
tiles_height = screen_height // tile_size
overhead_height = 3
max_drill_level = tiles_height - overhead_height
screen = pygame.display.set_mode((screen_width, screen_height))
clock = pygame.time.Clock()

pygame.key.set_repeat(500, 100)
pygame.display.set_caption('Drill Baby Drill!')
dt = 0

bearimage = pygame.image.load("img/bear.png")
boulderimage = pygame.image.load("img/boulder.png")
dirtimage = pygame.image.load("img/dirt.png")
drillbabyimage = pygame.image.load("img/drillbaby.png")
drillbitimage = pygame.image.load("img/drillbit.png")
drillshaftimage = pygame.image.load("img/drillshaft.png")
emptyholebottomimage = pygame.image.load("img/emptyholebottom.png")
shaftimage = pygame.image.load("img/shaft.png")
skyimage = pygame.image.load("img/sky.png")
surfaceimage = pygame.image.load("img/surface.png")
surfaceholeimage = pygame.image.load("img/surfacehole.png")
victoryimage = pygame.image.load("img/victory.png")

gamefont = pygame.font.Font("fonts/VT323-Regular.ttf", 24)
text_surface1 = gamefont.render("instruct: Use right/left to move baby and up/down to raise or lower your drill.",
                               False, pygame.Color('yellow'))
text_surface2 = gamefont.render("          find all the lost bears. don't drill into a rock. Win game.",
                               False, pygame.Color('yellow'))
flagfont = pygame.font.Font("fonts/VT323-Regular.ttf", 32)
flag_text_surface = flagfont.render("nope@nope.nope", False, pygame.Color('black'))
flag_message_text_surface1 = flagfont.render("You win! Drill Baby is reunited with", False, pygame.Color('yellow'))
flag_message_text_surface2 = flagfont.render("all its bears. Welcome to Flare-On 12.", False, pygame.Color('yellow'))

TileImages = {
    'dirt':dirtimage,
    'sky':skyimage,
    'surface':surfaceimage,
    'surfacehole':surfaceholeimage,
    'shaft':shaftimage,
    'emptyholebottom':emptyholebottomimage,
    'boulder':boulderimage
}

LevelNames = [
    'California',
    'Ohio',
    'Death Valley',
    'Mexico',
    'The Grand Canyon'
]

class BackgroundTile(pygame.sprite.Sprite):
    def __init__(self, x, y, type):
        super().__init__()
        self.setType(type)
        self.rect = self.image.get_rect()
        self.x = x
        self.y = y
        self.rect.top = self.y * tile_size
        self.rect.left = self.x * tile_size

    def setType(self, type):
        self.type = type
        self.image = TileImages[self.type]

    def draw(self, surface):
        surface.blit(self.image, self.rect)


class Drill:
    def __init__(self, max_levels):
        self.drill_level = 0
        self.max_levels = max_levels

    def retract(self):
        if self.drill_level > 0:
            self.drill_level -= 1
    
    def drill(self):
        if self.drill_level < self.max_levels:
            self.drill_level += 1

    def drillEngaged(self):
        return self.drill_level != 0
    
    def draw(self, x, y, surface):
        remaining_tiles = self.drill_level
        
        while remaining_tiles > 0:
            drilly = y + remaining_tiles
            if remaining_tiles == self.drill_level:
                drill_image = drillbitimage
            else:
                drill_image = drillshaftimage
            newrect = drill_image.get_rect()
            newrect.top = drilly * tile_size
            newrect.left = x * tile_size
            surface.blit(drill_image, newrect)
            remaining_tiles -= 1


class DrillBaby(pygame.sprite.Sprite):
    def __init__(self, x, y, max_levels):
        super().__init__()
        self.drill = Drill(max_levels)
        self.image = drillbabyimage
        self.rect = self.image.get_rect()
        self.x = x
        self.y = y
        self.rect.top = self.y * tile_size
        self.rect.left = self.x * tile_size

    def draw(self, surface):
        surface.blit(self.image, self.rect)
        self.drill.draw(self.x, self.y, surface)

    def move(self, dx):
        self.x += dx
        self.rect.left = self.x * tile_size

    def hitBoulder(self):
        global boulder_layout
        boulder_level = boulder_layout[self.x]
        return boulder_level == self.drill.drill_level
    
    def hitBear(self):
        return self.drill.drill_level == max_drill_level

background_tiles = []
boulder_layout = []


def AttemptPlayerMove(dx, dy):
    newx = player.x + dx

    # Can only move within screen bounds
    if newx < 0 or newx >= tiles_width:
        return False

    # Can only move side to side when drill is not engaged
    if dx != 0 and player.drill.drillEngaged():
        return False

    # Operate drill if they moved up or down
    if dy < 0:
        player.drill.retract()
    elif dy > 0:
        DrillTile()

    player.move(dx)
    return True


def DrillTile():
    global player
    player.drill.drill()
    global background_tiles    

    background_tile = background_tiles[player.y + player.drill.drill_level][player.x]

    if background_tile.type == 'dirt':
        if player.hitBoulder():
            background_tile.setType('boulder')
        else:
            background_tile.setType('emptyholebottom')

        tile_above = background_tiles[player.y + player.drill.drill_level - 1][player.x]
        
        if player.drill.drill_level == 1:
            tile_above.setType('surfacehole')
        if player.drill.drill_level > 1:
            tile_above.setType('shaft')


def GenerateFlagText(sum):
    key = sum >> 8
    encoded = "\xd0\xc7\xdf\xdb\xd4\xd0\xd4\xdc\xe3\xdb\xd1\xcd\x9f\xb5\xa7\xa7\xa0\xac\xa3\xb4\x88\xaf\xa6\xaa\xbe\xa8\xe3\xa0\xbe\xff\xb1\xbc\xb9"
    plaintext = []
    for i in range(0, len(encoded)):
        plaintext.append(chr(ord(encoded[i]) ^ (key+i)))
    return ''.join(plaintext)


def main():
    global background_tiles  
    global player
    global LevelNames
    global boulder_layout
    victory_mode = False
    bear_mode = False
    next_level_mode = False
    boulder_mode = False
    bear_sum = 1
    running = True
    current_level = 0
    flag_text = None

    random.shuffle(LevelNames)

    while running:
        background_tiles = BuildBackground()
        player = DrillBaby(7, 2, max_drill_level)
        boulder_layout = []
        for i in range(0, tiles_width):
            if (i != len(LevelNames[current_level])):
                boulder_layout.append(random.randint(2, max_drill_level))
            else:
                print(i)
                boulder_layout.append(-1)

        while running and not next_level_mode:

            # poll for events
            # pygame.QUIT event means the user clicked X to close your window
            for event in pygame.event.get():
                if event.type == pygame.QUIT:
                    running = False
                if event.type == pygame.KEYDOWN and not (boulder_mode or victory_mode):
                    if bear_mode:
                        bear_mode = False
                        next_level_mode = True
                    elif event.key == pygame.K_w or event.key == pygame.K_UP:
                        AttemptPlayerMove(0, -1)
                    elif event.key == pygame.K_s or event.key == pygame.K_DOWN:
                        AttemptPlayerMove(0, 1)
                    elif event.key == pygame.K_a or event.key == pygame.K_LEFT:
                        AttemptPlayerMove(-1, 0)
                    elif event.key == pygame.K_d or event.key == pygame.K_RIGHT:
                        AttemptPlayerMove(1, 0)

            #draw the background tiles
            for tilerow in background_tiles:
                for tile in tilerow:
                    tile.draw(screen)

            # display the instructions
            screen.blit(text_surface1, (0, 0))
            screen.blit(text_surface2, (0, 24))

            # display status message
            if boulder_mode:
                statustext = "You Hit a Boulder, please reload try again when you are better at video games"
                statussurface = gamefont.render(statustext, False, pygame.Color('red'))
                screen.blit(statussurface, (0, 48))
            else:
                if bear_mode:
                    statustext = "You found a bear, press any key to drill for another!"
                    statussurface = gamefont.render(statustext, False, pygame.Color('green'))
                    screen.blit(statussurface, (0, 48))
                
                # display stage name
                leveltext = 'Level: ' + LevelNames[current_level]
                levelnamesurface = gamefont.render(leveltext, False, pygame.Color('green'))
                screen.blit(levelnamesurface, (screen_width - 12*len(leveltext), 48))

            # display location info
            locationtext = "Loc: %d" % player.x
            depthtext = "Depth: %d" % player.drill.drill_level
            locationtextsurface = gamefont.render(locationtext, False, pygame.Color('cyan'))
            depthtextsurface = gamefont.render(depthtext, False, pygame.Color('cyan'))
            screen.blit(locationtextsurface, (700, 72))
            screen.blit(depthtextsurface, (700, 96))

            # draw the baby
            player.draw(screen)

            if player.hitBoulder():
                boulder_mode = True

            if player.hitBear():
                player.drill.retract()
                bear_sum *= player.x
                bear_mode = True

            if bear_mode:
                screen.blit(bearimage, (player.rect.x, screen_height - tile_size))
                if current_level == len(LevelNames) - 1 and not victory_mode:
                    victory_mode = True
                    flag_text = GenerateFlagText(bear_sum)
                    print("Your Flag: " + flag_text)

            if victory_mode:
                flag_text_surface = flagfont.render(flag_text, False, pygame.Color('black'))
                screen.blit(victoryimage, (111, 50))
                screen.blit(flag_message_text_surface1, (150, 60))
                screen.blit(flag_message_text_surface2, (150, 92))
                screen.blit(flag_text_surface, (200, 125))

            # flip() the display to put your work on screen
            pygame.display.flip()

            # limits FPS to 60
            # dt is delta time in seconds since last frame, used for framerate-
            # independent physics.
            dt = clock.tick(60) / 1000
        
        next_level_mode = False
        current_level += 1

    pygame.quit()
    return

def BuildBackground():
    tileset = []
    # make two rows of pure sky, one surface row, then fill the remaining rows with dirt
    for y in range(0, tiles_height):
        row = []
        for x in range(0, tiles_width):
            if y < 2:
                type = 'sky'
            elif y == 2:
                type = 'surface'
            else:
                type = 'dirt'
            
            row.append(BackgroundTile(x, y, type))
        tileset.append(row)

    return tileset

if __name__ == '__main__':
    main()