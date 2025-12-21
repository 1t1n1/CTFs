from sympy import symbols, Eq, solve

# Define the variables
mot5, mot4, mot3, mot2, mot1, mot0, mot6 = symbols('mot5 mot4 mot3 mot2 mot1 mot0 mot6')

# Define the equations
eq1 = Eq(mot5 + mot4 + mot3 + mot2 + mot1 + mot0 + mot6, 546)
eq2 = Eq(mot5 + mot4 + mot3 + mot2 + mot1 + mot0 - mot6, 480)
eq3 = Eq(mot4 + mot3 + mot2 + mot1 + mot0 - mot5 + mot6, 412)
eq4 = Eq(mot5 + mot3 + mot2 + mot1 + mot0 - mot4 + mot6, 440)
eq5 = Eq(mot5 + mot4 + mot2 + mot1 + mot0 - mot3 + mot6, 312)
eq6 = Eq(mot5 + mot4 + mot3 + mot1 + mot0 - mot2 + mot6, 356)
eq7 = Eq(mot5 + mot4 + mot3 + mot2 + mot0 - mot1 + mot6, 314)

# Solve the system of equations
solution = solve((eq1, eq2, eq3, eq4, eq5, eq6, eq7), (mot5, mot4, mot3, mot2, mot1, mot0, mot6))
print("".join([chr(mot) for mot in solution.values()]))
