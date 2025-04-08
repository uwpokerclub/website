import { Link, useParams } from "react-router-dom";
import { useFetch } from "../../../hooks";
import { APIErrorResponse, Entry, Event, StructureWithBlinds } from "../../../types";
import { useCallback, useEffect, useRef, useState } from "react";
import { sendAPIRequest } from "../../../lib";
import { EntriesTable } from "./EntriesTable";

import styles from "./EventDetails.module.css";
import { TournamentClock } from "../../tournament-clock";
import { EndEventModal } from "./EndEventModal";

export function EventDetails() {
  const [showModal, setShowModal] = useState(false);
  const { eventId = "" } = useParams<{ eventId: string }>();

  const { data: event, setData: setEvent, isLoading } = useFetch<Event>(`events/${eventId}`);
  const { data: entries, setData: setEntries } = useFetch<Entry[]>(`participants?eventId=${eventId}`);

  const [structure, setStructure] = useState<StructureWithBlinds>();
  const [showInfo, setShowInfo] = useState(true);
  const [showEntries, setShowEntries] = useState(true);
  const [error, setError] = useState("");

  const endEventBtnRef = useRef<HTMLButtonElement | null>(null);
  const restartEventBtnRef = useRef<HTMLButtonElement | null>(null);

  const updateParticipants = useCallback(async () => {
    const { status, data } = await sendAPIRequest<Entry[]>(`participants?eventId=${eventId}`);

    if (status === 200 && data) {
      setEntries(data);
    }
  }, [eventId, setEntries]);

  const handleModalSuccess = useCallback(() => {
    setEvent(
      event
        ? {
            ...event,
            state: 1,
          }
        : undefined,
    );

    updateParticipants();

    // Re-enable the end event button
    endEventBtnRef.current!.disabled = false;
  }, [event, setEvent, updateParticipants]);

  const handleModalClose = useCallback(() => {
    // Close the modal
    setShowModal(false);

    // Re-enable the end event button
    endEventBtnRef.current!.disabled = false;
  }, []);

  const handleRestartEvent = useCallback(async () => {
    // Disable restart event button while data is processing
    restartEventBtnRef.current!.disabled = true;

    const { status, data } = await sendAPIRequest<APIErrorResponse>(`events/${eventId}/unend`, "POST");

    if (data && status !== 204) {
      return setError(data.message);
    }

    setEvent((prev) => (prev ? { ...prev, state: 0 } : prev));

    updateParticipants();

    // Re-enable restart event button
    restartEventBtnRef.current!.disabled = false;
  }, [eventId, setEvent, updateParticipants]);

  const handleRebuy = async () => {
    const { status } = await sendAPIRequest(`events/${eventId}/rebuy`, "POST");
    if (status === 200) {
      setEvent((e) => {
        if (!e) return;

        return {
          ...e,
          rebuys: e.rebuys + 1,
        };
      });
    }
  };

  const handleEndEventClick = useCallback(() => {
    // Show the modal
    setShowModal(true);

    // Disable the button to prevent it from being clicked again
    endEventBtnRef.current!.disabled = true;
  }, []);

  useEffect(() => {
    if (!event) {
      return;
    }

    sendAPIRequest<StructureWithBlinds>(`structures/${event.structureId}`).then(({ data }) => {
      if (!data) {
        return;
      }
      setStructure(data);
    });
  }, [event]);

  return (
    <>
      {event !== undefined && !isLoading && (
        <div>
          {event.state === 1 && (
            <div data-qa="event-ended-banner" className="alert alert-danger">
              This event has ended.
            </div>
          )}
          {error && <div className="alert alert-danger">{error}</div>}

          <section className={styles.tabGroup}>
            <div
              data-qa="tournament-tab"
              onClick={() => setShowInfo(true)}
              className={`${styles.tab} ${showInfo ? styles.tabActive : ""}`}
            >
              Tournament Info
            </div>
            <div
              data-qa="clock-tab"
              onClick={() => setShowInfo(false)}
              className={`${styles.tab} ${!showInfo ? styles.tabActive : ""}`}
            >
              Clock
            </div>
          </section>
          {showInfo ? (
            <>
              <header className={styles.header}>
                <section className={styles.info}>
                  <h1>{event.name}</h1>
                  <p>
                    <strong>Format:</strong> {event.format}
                  </p>
                  <p>
                    <strong>Date:</strong>{" "}
                    {new Date(event.startDate).toLocaleDateString("en-US", {
                      weekday: "long",
                      year: "numeric",
                      month: "long",
                      day: "numeric",
                      hour: "numeric",
                      minute: "numeric",
                    })}
                  </p>
                  {event.notes && (
                    <p>
                      <strong>Additional Details:</strong> {event.notes}
                    </p>
                  )}
                </section>
                <section className={styles.actions}>
                  {event.state !== 1 && (
                    <>
                      <Link data-qa="register-members-btn" to={`register`} className="btn btn-primary">
                        Register Members
                      </Link>

                      <button data-qa="rebuy-btn" onClick={handleRebuy} type="button" className="btn btn-success">
                        Rebuy
                      </button>

                      <button
                        ref={endEventBtnRef}
                        data-qa="end-event-btn"
                        onClick={handleEndEventClick}
                        type="button"
                        className="btn btn-danger"
                      >
                        End Event
                      </button>
                    </>
                  )}
                  {event.state === 1 && (
                    <>
                      <button
                        ref={restartEventBtnRef}
                        data-qa="restart-event-btn"
                        onClick={handleRestartEvent}
                        type="button"
                        className="btn btn-danger"
                      >
                        Restart Event
                      </button>
                    </>
                  )}
                </section>
              </header>

              <section className={styles.tabGroup}>
                <div
                  onClick={() => setShowEntries(true)}
                  className={`${styles.tab} ${showEntries ? styles.tabActive : ""}`}
                >
                  Entries
                </div>
                <div
                  onClick={() => setShowEntries(false)}
                  className={`${styles.tab} ${!showEntries ? styles.tabActive : ""}`}
                >
                  Structure
                </div>
              </section>

              {showEntries ? (
                <EntriesTable entries={entries || []} event={event} updateParticipants={updateParticipants} />
              ) : (
                <>
                  <input
                    name="structureName"
                    className="form-control"
                    placeholder="Name the structure"
                    value={structure?.name || "Unknown Structure"}
                    disabled
                  />
                  <header className={styles.blindsHeader}>
                    <span>Small</span>
                    <span>Big</span>
                    <span>Ante</span>
                    <span>Time</span>
                  </header>
                  {structure?.blinds.map((blind, i) => (
                    <div key={i} className="input-group">
                      <input className="form-control" type="text" value={blind.small} disabled></input>
                      <input className="form-control" type="text" value={blind.big} disabled></input>
                      <input className="form-control" type="text" value={blind.ante} disabled></input>
                      <input className="form-control" type="text" value={blind.time} disabled></input>
                    </div>
                  )) || <></>}
                </>
              )}
            </>
          ) : (
            <TournamentClock levels={structure?.blinds || []} />
          )}
        </div>
      )}
      <EndEventModal show={showModal} onClose={handleModalClose} onSuccess={handleModalSuccess} />
    </>
  );
}
