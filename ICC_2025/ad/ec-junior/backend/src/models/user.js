const db = require('../config/database');
const { v4: uuidv4 } = require('uuid');

class User {
    constructor(data = {}) {
        this.id = data.id ?? null;
        this.email = data.email ?? '';
        this.passwordHash = data.passwordHash ?? '';
        this.firstName = data.firstName ?? '';
        this.lastName = data.lastName ?? '';
        this.isAdmin = data.isAdmin ?? false;
        this.creditBalance = data.creditBalance ?? 1000.00;
        this.createdAt = data.createdAt ?? new Date();
    }

    static async findByEmail({ email }) {
        const row = await db.get('SELECT * FROM users WHERE email = ?', [email]);
        return row ? new User(row) : null;
    }

    static async findById({ id }) {
        const row = await db.get('SELECT * FROM users WHERE id = ?', [id]);
        return row ? new User(row) : null;
    }

    static async create(userData) {
        const id = uuidv4();
        await db.run(
            'INSERT INTO users (id, email, passwordHash, firstName, lastName, isAdmin, creditBalance) VALUES (?, ?, ?, ?, ?, ?, ?)',
            [
                id,
                userData.email,
                userData.passwordHash,
                userData.firstName || '',
                userData.lastName || '',
                userData.isAdmin || false,
                userData.creditBalance || 1000.00
            ]
        );
        return new User({ id, ...userData });
    }

    static async updateById({ id, ...userData }) {
        await db.run(
            'UPDATE users SET firstName = ?, lastName = ?, email = ? WHERE id = ?',
            [userData.firstName || '', userData.lastName || '', userData.email, id]
        );
        return await User.findById({ id });
    }

    static async decrementCreditBalance({ userId, amount }) {
        await db.run(
            'UPDATE users SET creditBalance = creditBalance - ? WHERE id = ?',
            [amount, userId]
        );
    }

    toJSON() {
        return {
            id: this.id,
            email: this.email,
            firstName: this.firstName,
            lastName: this.lastName,
            isAdmin: this.isAdmin,
            creditBalance: this.creditBalance,
            createdAt: this.createdAt
        };
    }
}

module.exports = User;