# ymm5 commence à 0

# ymm5[i] = chiffre[i] + ymm5[i-1] * 10

final_val = 0xcf5bceab # 3478900395 -> 5930098743: Pas bon, ça amène à  0x61762037
# 0xcf5bceab = chiffre[n] + (chiffre[n-1] + (chiffre[n-2] + ymm5[n-3] * 10) * 10) * 10

# 1. Identifier les syscalls
# 2. Mais pourquoi est-ce que je trouve pas le code qui me demande mon nom et code de registration?
# 3. En isolant, je vois que c'est le parent qui doit print, mais il casse quand je débug à cause du cmovq (comprendre pourquoi ça casse pas quand je débug pas)
# 4. Passer par dessus ce problème en jumpant, puis voir que c'est bien du code exécutable
# 5. Objdump le code pcq ghidra y arrive pas
# 6. Analyser le nouveau code pour voir que y'a une partie qui traite le code, suivi d'une partie qui traite le nom
# 7. La partie qui traite le nom est inutile à comprendre (autre que savoir que le résultat va dans xmm0) 
# 8. La partie qui traite le code doit être comprise. Elle place le résultat dans xmm5
# 9. xmm1 est constant et m'est inconnu
# 10. Au final les trois sont xors ensemble pour donner 0
# 11. On retourne sur la partie qui traite le code et on voit que ça fait juste mettre en int. On reverse le tout et on réussit!

# Note: J'ai aussi décodé les strings pour m'aider à comprendre