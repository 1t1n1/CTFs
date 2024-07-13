# 02 - Pixel Poker
## Description
I said you wouldn't win that last one. I lied. The last challenge was basically a captcha. Now the real work begins. Shall we play another game?

## Fichiers
PixelPoker.exe

---

# The Battle

### Interesting strings
- GetPixel
- SetPixel
- "Womp womp... :("

FUN_004012c0 semble être celle qui s'occupe de la vérification des clicks (elle affiche "Womp womp... :( après 10 clicks)

if ((x_coord == s_FLARE-On_00412004._0_4_ % DAT_00413280) &&
    (y_coord == s_FLARE-On_00412004._4_4_ % DAT_00413284)) {...

Cette ligne semble d'ailleurs s'occuper de cette vérification
Avec s_FLARE-On_00412004 = "FLARE-On" (en mémoire), on trouve 
s_FLARE-On_00412004._0_4_ == 0x46
s_FLARE-On_00412004._4_4_ == 0x45

Test 1 : Essayer la coord (70, 69) (0x46 = 70 et 0x45 = 69)
Résultat : "Womp womp... :("

Solution: Trouver la valeur de DAT_00413280 et DAT_00413284.
Comment faire? Je vais essayer d'utiliser un debugger et de trouver la valeur du registre dynamiquement en placant un breakpoint à l'adresse 0x00401482 et regarder la valeur de ESI qui devrait contenir la valeur du diviseur DAT_00413280.
Aussi, rendu là, regarder la valeur de EDX directement!
On obtient:
x_coord = 0x5F
y_coord = 0x139

Et on obtient le flag!!

Réponse: w1nN3r_W!NneR_cHick3n_d1nNer@flare-on.com