package src;

class Addon extends Item {
    public Addon(String name) {
        super(name, 0, 1);
    }

    public int calculatePrice() {
        switch (this.name) {
        case "cream":
            return 30;
        case "milk":
            return 20;
        default:
            return 10;
        }
    }
}
