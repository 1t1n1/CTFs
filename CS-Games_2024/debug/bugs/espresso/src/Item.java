package src;

class Item {
    public String name;
    public int price;
    public int quantity;

    public Item(String name, int price, int quantity) {
        this.name = name;
        this.price = price;
        this.quantity = quantity;
    }

    public int calculatePrice() {
        return this.price * this.quantity;
    }
}
