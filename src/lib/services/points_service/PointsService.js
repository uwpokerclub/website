const pointsTable = require("./point_lookup_table.json");

const SIZE_FACTOR = 50;

class PointsService {
  constructor(size) {
    this.size = size;
  }

  calculatePoints(placement) {
    const points = pointsTable[`${placement}`] || 1;
    return Math.floor((points * this.size) / SIZE_FACTOR);
  }
}

module.exports = PointsService;
