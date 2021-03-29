import * as pointsTable from "./point_lookup_table.json";

const SIZE_FACTOR = 50;

type PointsTable = {
  [key: string]: number;
};

export default class PointsService {
  private size: number;

  constructor(size: number) {
    this.size = size;
  }

  public calculatePoints(placement: number): number {
    const points = (pointsTable as PointsTable)[placement.toString()] || 1;
    return Math.floor((points * this.size) / SIZE_FACTOR);
  }
}
