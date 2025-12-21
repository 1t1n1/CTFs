#include <iostream>

using namespace std;

template<int N> requires (!(N % 2) or !(N % 3) or !(N % 5) or !(N % 7) or !(N % 11))
constexpr bool _17d5021c5c78985f6b27df452366e353()
{
    return false;
}
template<int N>
constexpr bool _17d5021c5c78985f6b27df452366e353()
{
    return true;
}

template<int N>
struct _9aeaed51f2b0f6680c4ed4b07fb1a83c
{
    static const int value = _9aeaed51f2b0f6680c4ed4b07fb1a83c<N - 1>::value;
};

template<int F> requires (F % 5 != 0)
constexpr int gimmeThatFlag()
{
    return _9aeaed51f2b0f6680c4ed4b07fb1a83c<F>::value;
}

template<int F>
constexpr int gimmeThatFlag()
{
    return F % 7;
}

template<int Z> requires (sizeof(int) == 4 and !((Z >> 31) & 1))
constexpr int comOnIJustWantThatFlag()
{
    return _9aeaed51f2b0f6680c4ed4b07fb1a83c<Z>::value;
}

template<int Z>
constexpr int comOnIJustWantThatFlag()
{
    return 15;
}


template<int X, int N>
struct OSS {
    static const int value = OSS<X+15, N+1>::value;
};

template<int N>
struct OSS<1755, N> {
    static const int value = N;
};


template<int N> requires (N == 0) constexpr int iDontWantToPlayAnymoreLetMeOut();
template<int N> requires ((N % 10) < 2 && N > 0) constexpr int iDontWantToPlayAnymoreLetMeOut();
template<int N> requires (N == 7) constexpr int iDontWantToPlayAnymoreLetMeOut();
template<int N> requires (N == 3) constexpr int iDontWantToPlayAnymoreLetMeOut();
template<int N> constexpr int iDontWantToPlayAnymoreLetMeOut();

template<int N> requires (N == 0)
constexpr int iDontWantToPlayAnymoreLetMeOut()
{
    return N;
}

template<int N> requires ((N % 10) < 2 && N > 0)
constexpr int iDontWantToPlayAnymoreLetMeOut()
{
    return N % 10 + 2 * iDontWantToPlayAnymoreLetMeOut<N / 10>();
}

template<int N> requires (N == 7)
constexpr int iDontWantToPlayAnymoreLetMeOut()
{
    return iDontWantToPlayAnymoreLetMeOut<N - 4>();
}

template<int N> requires (N == 3)
constexpr int iDontWantToPlayAnymoreLetMeOut()
{
    return iDontWantToPlayAnymoreLetMeOut<N + 4>();
}

template<int N>
constexpr int iDontWantToPlayAnymoreLetMeOut()
{
    return iDontWantToPlayAnymoreLetMeOut<N * 0 + N - N + 3>();
}

template <int x, int t> requires (x == 0)
constexpr int iSaidLetMeOuuuuuuut()
{
    return t;
}

template <int x, int t>
constexpr int iSaidLetMeOuuuuuuut()
{
    return iSaidLetMeOuuuuuuut<x / 2, t * 2 + x % 2>();
}

template <int x>
constexpr int iSaidLetMeOuuuuuuut()
{
    return iSaidLetMeOuuuuuuut<x, 0>();
}

template<
    int Conçu, int par, int l_esprit, int diabolique, int du, int Baron, int Sukumvit,
    int LE, int LABYRINTHE, int DE, int LA, int MORT,
    int est, int truffé, int de,
    int pièges, int mortels, int et, int peuplé,
    int _de, int monstres, int assoiffés, int de_, int sang,

    int ET, int VOUS, int oserez, int vous, int y, int entrer> // ?
requires (
    -~de == pièges and
    (par % 2 == 0) and
    (par % 6 == 4) and
    iDontWantToPlayAnymoreLetMeOut<et>() == 6 and
    oserez == OSS<0, 0>::value and
    (LE / 20 == LE % 10) and
    _17d5021c5c78985f6b27df452366e353<Baron>() and
    est % 10 == est / 10 and
    (du * 2 + 7 == MORT) and
    (peuplé == ((peuplé / 10) * 2 * 5)) and
    (80 / (LABYRINTHE - Baron) == 10) and
    !(assoiffés - peuplé + ~0) and
    iDontWantToPlayAnymoreLetMeOut<monstres + 994>() == 15 and
    Baron % 7 == 3 and
    DE % 14 == 0 and
    de == entrer and
    !_17d5021c5c78985f6b27df452366e353<vous - 5>() and
    (assoiffés + de - de_ - est == monstres - vous) and
    (diabolique - Conçu == LABYRINTHE - _de) and
    (mortels == Baron) and
    (char)(truffé - 212) == 127 and
    ((LABYRINTHE - LA) == 1 - ((LABYRINTHE - LA) != (LABYRINTHE - LA))) and
    gimmeThatFlag<Sukumvit>() == 3 and
    iDontWantToPlayAnymoreLetMeOut<y>() == 7 and
    LE < Baron and
    (vous / 58) == (int) ((float)vous / 58.0f) and
    iSaidLetMeOuuuuuuut<VOUS>() == 67 and
    ~de == -115 and
    (!!(diabolique - Conçu) == (diabolique - Conçu)) and
    (Conçu * par * l_esprit == 345800) and
    (Conçu % par == 70) and
    vous != 42 and
    Sukumvit > Baron and
    (du % comOnIJustWantThatFlag<128 - du * 3>() == 0) and
    (par % 19 == 0) and
    Baron % 9 == 2 and
    LE % 4 == !!_17d5021c5c78985f6b27df452366e353<LE>() and
    iDontWantToPlayAnymoreLetMeOut<ET>() == 6 and
    pièges - sang == peuplé - Conçu and
    (est / 7) % 10 == (est / 7) / 10 and
    comOnIJustWantThatFlag<LABYRINTHE - DE>() % 3 == 0 and
    (peuplé == ((peuplé / 8) * 2 * 4)) and
    -diabolique == Baron - diabolique + est - (est + Baron)
)
class Flag
{
public:
    void show_flag() const {

        cout << (char) Conçu;
        cout << (char) par;
        cout << (char) l_esprit;
        cout << (char) diabolique;
        cout << (char) du;
        cout << (char) par;
        cout << (char) Baron;
        cout << (char) Sukumvit;
        cout << (char) LE;
        cout << (char) Baron;
        cout << (char) LABYRINTHE;
        cout << (char) DE;
        cout << (char) LA;
        cout << (char) MORT;
        cout << (char) vous;
        cout << (char) Baron;
        cout << (char) Sukumvit;
        cout << (char) est;
        cout << (char) Baron;
        cout << (char) truffé;
        cout << (char) vous;
        cout << (char) de;
        cout << (char) Baron;
        cout << (char) pièges;
        cout << (char) Sukumvit;
        cout << (char) mortels;
        cout << (char) et;
        cout << (char) vous;
        cout << (char) peuplé;
        cout << (char) _de;
        cout << (char) monstres;
        cout << (char) Sukumvit;
        cout << (char) assoiffés;
        cout << (char) de_;
        cout << (char) Baron;
        cout << (char) par;
        cout << (char) Baron;
        cout << (char) est;
        cout << (char) sang;
        cout << (char) ET;
        cout << (char) y;
        cout << (char) vous;
        cout << (char) VOUS;
        cout << (char) oserez;
        cout << (char) entrer;
        cout << (char) Baron;

        cout << endl;
    }
};

int main() {
    int Conçu;
    int par;
    int l_esprit;
    int diabolique;
    int du;
    int Baron;
    int Sukumvit;
    int LE;
    int LABYRINTHE;
    int DE;
    int LA;
    int MORT;
    int est;
    int truffé;
    int de;
    int pièges;
    int mortels;
    int et;
    int peuplé;
    int _de;
    int monstres;
    int assoiffés;
    int de_;
    int sang;
    int ET;
    int VOUS;
    int oserez;
    int vous;
    int y;
    int entrer;
    Flag<Conçu, par, l_esprit, diabolique, du, Baron, Sukumvit, LE, LABYRINTHE, DE, LA, MORT, est, truffé, de, pièges, mortels, et, peuplé, _de, monstres, assoiffés, de_, sang, ET, VOUS, oserez, vous, y, entrer> f;
    f.show_flag();
}
