import * as pointsTable from "./point_lookup_table.json";

const SIZE_FACTOR = 50;

class PointsService {
  private size: number;

  constructor(size: number) {
    this.size = size;
  }

  calculatePoints(placement) {
    const points = pointsTable[`${placement}`] || 1;
    return Math.floor((points * this.size) / SIZE_FACTOR);
  }
}

module.exports = PointsService;
