const crypto = require('crypto');
const db = require('../config/database');

class Inquiry {
    constructor(data = {}) {
        this.id = data.id ?? null;
        this.userId = data.userId ?? '';
        this.subject = data.subject ?? '';
        this.message = data.message ?? '';
        this.inquiryType = data.inquiryType ?? 'general';
        this.productId = data.productId ?? null;
        this.response = data.response ?? null;
        this.respondedAt = data.respondedAt ?? null;
        this.createdAt = data.createdAt ?? new Date();
    }

    static async findById({ id }) {
        const row = await db.get('SELECT * FROM inquiries WHERE id = ?', [id]);
        return row ? new Inquiry(row) : null;
    }

    static async findByUserId({ userId }) {
        const rows = await db.all(
            'SELECT * FROM inquiries WHERE userId = ? ORDER BY createdAt DESC',
            [userId]
        );
        return rows.map(row => new Inquiry(row));
    }

    static async findAll() {
        const rows = await db.all('SELECT * FROM inquiries ORDER BY createdAt DESC');
        return rows.map(row => new Inquiry(row));
    }

    static async create(inquiryData) {
        const id = crypto.randomBytes(4).toString('hex');;
        await db.run(
            'INSERT INTO inquiries (id, userId, subject, message, inquiryType, productId) VALUES (?, ?, ?, ?, ?, ?)',
            [id, inquiryData.userId, inquiryData.subject, inquiryData.message, inquiryData.inquiryType || 'general', inquiryData.productId || null]
        );
        return new Inquiry({ id, ...inquiryData });
    }

    static async addResponse({ id, response }) {
        await db.run(
            'UPDATE inquiries SET response = ?, respondedAt = CURRENT_TIMESTAMP WHERE id = ?',
            [response, id]
        );
        return await Inquiry.findById({ id });
    }

    toJSON() {
        return {
            id: this.id,
            userId: this.userId,
            subject: this.subject,
            message: this.message,
            createdAt: this.createdAt
        };
    }
}

module.exports = Inquiry;
