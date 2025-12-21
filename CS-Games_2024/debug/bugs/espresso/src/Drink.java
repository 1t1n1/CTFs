package src;

import java.util.List;

class Drink extends Item {
    public String size;
    public List<Addon> addons;

    public Drink(String name, int price, String size, int quantity) {
        super(name, price, quantity);
        this.size = size;
    }

    public int calculatePrice() {
        int price = this.price;

        switch (this.size) {
            case "medium":
                price += 100;
                break;
            case "large":
                price += 200;
                break;
        }

        for (Addon addon : this.addons) {
            price += addon.price;
        }

        return price * this.quantity;
    }
}
