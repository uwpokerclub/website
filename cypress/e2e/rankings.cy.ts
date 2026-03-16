import { RANKINGS, SEMESTER } from "../seed";
import { getUserForMember } from "../support/helpers";

// Pre-computed sorted rankings with user info for easier test assertions
// Note: The API returns user ID as the ranking id, not membershipId
const SORTED_RANKINGS = RANKINGS.map((r) => {
  const user = getUserForMember(r.membershipId);
  return {
    ...r,
    user,
    id: user.id, // API returns user ID as ranking.id
  };
}).sort((a, b) => {
  if (b.points !== a.points) return b.points - a.points;
  const lastCmp = a.user.lastName.localeCompare(b.user.lastName);
  if (lastCmp !== 0) return lastCmp;
  return a.user.firstName.localeCompare(b.user.firstName);
});

describe("Rankings", () => {
  before(() => {
    cy.resetDatabase();
  });

  context("when no semester is selected", () => {
    beforeEach(() => {
      cy.login();
      // Mock semesters API to return empty array so no semester is selected
      cy.intercept("GET", "/api/v2/semesters", { data: [], total: 0 }).as("getSemesters");
    });

    it("should display no semester selected message", () => {
      cy.visit("/admin/rankings");
      cy.getByData("rankings-no-semester").should("be.visible");
    });
  });

  context("loading state", () => {
    beforeEach(() => {
      cy.login();
      cy.intercept("GET", "/api/v2/semesters", { fixture: "semesters.json" }).as("getSemesters");
    });

    it("should display loading spinner while fetching rankings", () => {
      // Intercept with delay to see loading state
      cy.intercept("GET", /\/api\/v2\/semesters\/.*\/rankings(?:\?|$)/, {
        statusCode: 200,
        fixture: "rankings.json",
        delay: 500,
      }).as("slowRankings");

      cy.visit("/admin/rankings");
      cy.getByData("rankings-loading").should("be.visible");
      cy.wait("@slowRankings");
      cy.getByData("rankings-loading").should("not.exist");
    });
  });

  context("error state", () => {
    beforeEach(() => {
      cy.login();
      cy.intercept("GET", "/api/v2/semesters", { fixture: "semesters.json" }).as("getSemesters");
    });

    it("should display error message and retry button when API fails", () => {
      cy.intercept("GET", /\/api\/v2\/semesters\/.*\/rankings(?:\?|$)/, {
        statusCode: 500,
        body: { error: "Internal Server Error" },
      }).as("getRankingsError");

      cy.visit("/admin/rankings");
      cy.wait("@getRankingsError");

      cy.getByData("rankings-error").should("be.visible");
      cy.getByData("retry-btn").should("exist");
    });
  });

  context("with semester selected", () => {
    beforeEach(() => {
      cy.login();
      cy.intercept("GET", "/api/v2/semesters", { fixture: "semesters.json" }).as("getSemesters");
      cy.intercept("GET", /\/api\/v2\/semesters\/.*\/rankings(?:\?|$)/, { fixture: "rankings.json" }).as("getRankings");
      cy.visit("/admin/rankings");
      // Wait for table to be visible (data loaded)
      cy.getByData("rankings-table").should("exist");
    });

    context("podium display", () => {
      it("should display podium with correct data and layout", () => {
        cy.getByData("rankings-podium").should("be.visible");
        cy.getByData("podium-position-1").should("exist");
        cy.getByData("podium-position-2").should("exist");
        cy.getByData("podium-position-3").should("exist");

        // Correct names and points
        const first = SORTED_RANKINGS[0];
        const second = SORTED_RANKINGS[1];
        const third = SORTED_RANKINGS[2];

        cy.getByData("podium-name-1").should(
          "contain",
          `${first.user.firstName} ${first.user.lastName}`
        );
        cy.getByData("podium-points-1").should("contain", `${first.points} pts`);

        cy.getByData("podium-name-2").should(
          "contain",
          `${second.user.firstName} ${second.user.lastName}`
        );
        cy.getByData("podium-points-2").should("contain", `${second.points} pts`);

        cy.getByData("podium-name-3").should(
          "contain",
          `${third.user.firstName} ${third.user.lastName}`
        );
        cy.getByData("podium-points-3").should("contain", `${third.points} pts`);

        // Visual order: 2nd, 1st, 3rd
        cy.getByData("rankings-podium")
          .children()
          .first()
          .should("have.attr", "data-qa", "podium-position-2");
        cy.getByData("rankings-podium")
          .children()
          .eq(1)
          .should("have.attr", "data-qa", "podium-position-1");
        cy.getByData("rankings-podium")
          .children()
          .last()
          .should("have.attr", "data-qa", "podium-position-3");
      });

      it("should not display podium when fewer than 3 rankings", () => {
        // Mock response with only 2 rankings
        const twoRankings = [
          { id: "1", firstName: "User", lastName: "One", points: 10 },
          { id: "2", firstName: "User", lastName: "Two", points: 5 },
        ];

        cy.intercept("GET", /\/api\/v2\/semesters\/.*\/rankings(?:\?|$)/, {
          statusCode: 200,
          body: { data: twoRankings, total: twoRankings.length },
        }).as("getTwoRankings");

        cy.visit("/admin/rankings");
        cy.wait("@getTwoRankings");

        cy.getByData("rankings-podium").should("not.exist");
        cy.getByData("rankings-table").should("exist");
      });
    });

    context("table display", () => {
      it("should display table with correct data", () => {
        // Column headers
        cy.getByData("rank-header").should("contain", "Rank");
        cy.getByData("name-header").should("contain", "Name");
        cy.getByData("points-header").should("contain", "Points");

        // All rankings present
        SORTED_RANKINGS.forEach((ranking) => {
          cy.getByData(`ranking-row-${ranking.id}`).should("exist");
        });

        // First ranking data correct
        const first = SORTED_RANKINGS[0];
        cy.getByData(`ranking-rank-${first.id}`).should("contain", "1");
        cy.getByData(`ranking-name-${first.id}`).should(
          "contain",
          `${first.user.firstName} ${first.user.lastName}`
        );
        cy.getByData(`ranking-points-${first.id}`).should(
          "contain",
          first.points
        );

        // Results info
        cy.getByData("rankings-results-info").should(
          "contain",
          `Showing ${SORTED_RANKINGS.length} of ${SORTED_RANKINGS.length} rankings`
        );
      });
    });

    context("search functionality", () => {
      it("should filter rankings by name", () => {
        const searchUser = SORTED_RANKINGS[0].user;

        // By first name
        cy.getByData("input-rankings-search").type(searchUser.firstName);
        cy.getByData("rankings-results-info").should(
          "contain",
          searchUser.firstName
        );
        cy.getByData(`ranking-row-${SORTED_RANKINGS[0].id}`).should("exist");

        // Clear and search by last name
        cy.getByData("clear-search-btn").click();
        cy.getByData("input-rankings-search").type(searchUser.lastName);
        cy.getByData("rankings-results-info").should(
          "contain",
          searchUser.lastName
        );
        cy.getByData(`ranking-row-${SORTED_RANKINGS[0].id}`).should("exist");

        // Clear and search by full name
        cy.getByData("clear-search-btn").click();
        const fullName = `${searchUser.firstName} ${searchUser.lastName}`;
        cy.getByData("input-rankings-search").type(fullName);
        cy.getByData("rankings-results-info").should("contain", fullName);
        cy.getByData(`ranking-row-${SORTED_RANKINGS[0].id}`).should("exist");
      });

      it("should be case-insensitive and display search term in results", () => {
        const searchUser = SORTED_RANKINGS[0].user;
        const upperName = searchUser.firstName.toUpperCase();
        cy.getByData("input-rankings-search").type(upperName);

        cy.getByData("rankings-results-info").should("contain", upperName);
        cy.getByData(`ranking-row-${SORTED_RANKINGS[0].id}`).should("exist");

        // Verify search term display format
        cy.getByData("clear-search-btn").click();
        cy.getByData("input-rankings-search").type("Wald");
        cy.getByData("rankings-results-info").should(
          "contain",
          'matching "Wald"'
        );
      });

      it("should show no results state", () => {
        cy.getByData("input-rankings-search").type("nonexistentuser12345");

        cy.getByData("rankings-no-results").should("be.visible");
      });

      it("should clear search and show all rankings", () => {
        // First search to filter
        const searchUser = SORTED_RANKINGS[0].user;
        cy.getByData("input-rankings-search").type(searchUser.firstName);
        cy.getByData("rankings-results-info").should(
          "contain",
          searchUser.firstName
        );

        // Clear search
        cy.getByData("clear-search-btn").click();

        // Should show all rankings again
        cy.getByData("input-rankings-search").should("have.value", "");
        cy.getByData("rankings-results-info").should(
          "contain",
          `Showing ${SORTED_RANKINGS.length} of ${SORTED_RANKINGS.length}`
        );
      });
    });

    context("export functionality", () => {
      it("should handle export failure gracefully", () => {
        cy.intercept("GET", /\/api\/v2\/semesters\/.*\/rankings\/export/, {
          statusCode: 500,
          body: { error: "Export failed" },
        }).as("exportRankingsError");

        cy.getByData("export-rankings-btn").click();

        cy.wait("@exportRankingsError");
        // The button should still exist (not crash)
        cy.getByData("export-rankings-btn").should("exist");
      });
    });

    context("pagination", () => {
      it("should not show pagination when 25 or fewer rankings", () => {
        // With only 5 rankings in seed data, pagination should not be visible
        cy.getByData("rankings-pagination").should("not.exist");
      });

      it("should show pagination when more than 25 rankings", () => {
        // Generate mock data with 30 rankings
        const mockRankings = Array.from({ length: 30 }, (_, i) => ({
          id: `mock-${i}`,
          firstName: `User`,
          lastName: `${i + 1}`,
          points: 100 - i,
        }));

        cy.intercept("GET", /\/api\/v2\/semesters\/.*\/rankings(?:\?|$)/, {
          statusCode: 200,
          body: { data: mockRankings, total: mockRankings.length },
        }).as("getManyRankings");

        cy.visit("/admin/rankings");
        cy.wait("@getManyRankings");

        cy.getByData("rankings-pagination").should("exist");
      });

      it("should navigate to next page when clicking next", () => {
        // Generate mock data with 30 rankings
        const mockRankings = Array.from({ length: 30 }, (_, i) => ({
          id: `mock-${i}`,
          firstName: `User`,
          lastName: `${i + 1}`,
          points: 100 - i,
        }));

        cy.intercept(
          "GET",
          /\/api\/v2\/semesters\/.*\/rankings(?:\?|$)/,
          (req) => {
            const url = new URL(req.url);
            const offset = parseInt(
              url.searchParams.get("offset") || "0"
            );
            const limit = parseInt(
              url.searchParams.get("limit") || "25"
            );
            req.reply({
              statusCode: 200,
              body: {
                data: mockRankings.slice(offset, offset + limit),
                total: mockRankings.length,
              },
            });
          }
        ).as("getManyRankings");

        cy.visit("/admin/rankings");
        cy.wait("@getManyRankings");

        // Verify first page shows first 25
        cy.getByData("rankings-results-info").should("contain", "Showing 25");

        // Click next page
        cy.getByData("rankings-pagination").contains("button", "2").click();

        // Should now show remaining items
        cy.getByData("rankings-results-info").should("contain", "Showing 5");
      });
    });

    context("empty state", () => {
      it("should display empty state when no rankings exist", () => {
        cy.intercept("GET", /\/api\/v2\/semesters\/.*\/rankings(?:\?|$)/, {
          statusCode: 200,
          body: { data: [], total: 0 },
        }).as("getEmptyRankings");

        cy.visit("/admin/rankings");
        cy.wait("@getEmptyRankings");

        cy.getByData("rankings-empty").should("be.visible");
      });
    });
  });

  context("contract tests", () => {
    before(() => {
      cy.resetDatabase();
    });

    beforeEach(() => {
      cy.login();
    });

    it("should load rankings from real API", () => {
      cy.visit("/admin/rankings");
      cy.getByData("rankings-table").should("exist");

      SORTED_RANKINGS.forEach((ranking) => {
        cy.getByData(`ranking-row-${ranking.id}`).should("exist");
      });
    });

    it("should download CSV when export button clicked", () => {
      cy.intercept("GET", /\/api\/v2\/semesters\/.*\/rankings\/export/).as(
        "exportRankings"
      );

      cy.visit("/admin/rankings");
      cy.getByData("rankings-table").should("exist");

      cy.getByData("export-rankings-btn").click();

      cy.wait("@exportRankings")
        .its("response.statusCode")
        .should("eq", 200);
    });
  });
});
