import sqlite3
import os
import fastmcp
from datetime import date, datetime
from typing import Optional
from pydantic import BaseModel

import pytz
from fastmcp import Context


class YesNoSchemaModel(BaseModel):
    """A simple schema for yes/no questions."""
    answer: bool


db = os.environ.get("TRAVEL_DB", "tools/travel2.sqlite")
mcp = fastmcp.FastMCP()


@mcp.tool
async def fetch_user_flight_information(context: Context) -> list[dict] | str:
    """Fetch all tickets for the user along with corresponding flight information and seat assignments.

    Returns:
        A list of dictionaries where each dictionary contains the ticket details,
        associated flight details, and the seat assignments for each ticket belonging to the user.
    """

    passenger_id = get_passenger_id(context)
    conn = sqlite3.connect(db)
    cursor = conn.cursor()

    query = """
            SELECT t.ticket_no,
                   t.book_ref,
                   f.flight_id,
                   f.flight_no,
                   f.departure_airport,
                   f.arrival_airport,
                   f.scheduled_departure,
                   f.scheduled_arrival,
                   bp.seat_no,
                   tf.fare_conditions
            FROM tickets t
                     JOIN ticket_flights tf ON t.ticket_no = tf.ticket_no
                     JOIN flights f ON tf.flight_id = f.flight_id
                     JOIN boarding_passes bp ON bp.ticket_no = t.ticket_no AND bp.flight_id = f.flight_id
            WHERE t.passenger_id = ? \
            """
    cursor.execute(query, (passenger_id,))
    rows = cursor.fetchall()
    column_names = [column[0] for column in cursor.description]
    results = [dict(zip(column_names, row)) for row in rows]

    cursor.close()
    conn.close()

    return results


@mcp.tool
def search_flights(
        departure_airport: Optional[str] = None,
        arrival_airport: Optional[str] = None,
        start_time: Optional[date | datetime] = None,
        end_time: Optional[date | datetime] = None,
        limit: int = 20,
) -> list[dict]:
    """Search for flights based on departure airport, arrival airport, and departure time range."""
    conn = sqlite3.connect(db)
    cursor = conn.cursor()

    query = "SELECT * FROM flights WHERE 1 = 1"
    params = []

    if departure_airport:
        query += " AND departure_airport = ?"
        params.append(departure_airport)

    if arrival_airport:
        query += " AND arrival_airport = ?"
        params.append(arrival_airport)

    if start_time:
        query += " AND scheduled_departure >= ?"
        params.append(start_time)

    if end_time:
        query += " AND scheduled_departure <= ?"
        params.append(end_time)
    query += " LIMIT ?"
    params.append(limit)
    cursor.execute(query, params)
    rows = cursor.fetchall()
    column_names = [column[0] for column in cursor.description]
    results = [dict(zip(column_names, row)) for row in rows]

    cursor.close()
    conn.close()

    return results


@mcp.tool
def update_ticket_to_new_flight(context: Context, ticket_no: str, new_flight_id: int) -> str:
    """Update the user's ticket to a new valid flight."""

    passenger_id = get_passenger_id(context)
    conn = sqlite3.connect(db)
    cursor = conn.cursor()

    cursor.execute(
        "SELECT departure_airport, arrival_airport, scheduled_departure FROM flights WHERE flight_id = ?",
        (new_flight_id,),
    )
    new_flight = cursor.fetchone()
    if not new_flight:
        cursor.close()
        conn.close()
        return "Invalid new flight ID provided."
    column_names = [column[0] for column in cursor.description]
    new_flight_dict = dict(zip(column_names, new_flight))
    timezone = pytz.timezone("Etc/GMT-3")
    current_time = datetime.now(tz=timezone)
    departure_time = datetime.strptime(
        new_flight_dict["scheduled_departure"], "%Y-%m-%d %H:%M:%S.%f%z"
    )
    time_until = (departure_time - current_time).total_seconds()
    if time_until < (3 * 3600):
        return f"Not permitted to reschedule to a flight that is less than 3 hours from the current time. Selected flight is at {departure_time}."

    cursor.execute(
        "SELECT flight_id FROM ticket_flights WHERE ticket_no = ?", (ticket_no,)
    )
    current_flight = cursor.fetchone()
    if not current_flight:
        cursor.close()
        conn.close()
        return "No existing ticket found for the given ticket number."

    # Check the signed-in user actually has this ticket
    cursor.execute(
        "SELECT * FROM tickets WHERE ticket_no = ? AND passenger_id = ?",
        (ticket_no, passenger_id),
    )
    current_ticket = cursor.fetchone()
    if not current_ticket:
        cursor.close()
        conn.close()
        return f"Current signed-in passenger with ID {passenger_id} not the owner of ticket {ticket_no}"

    user_response = context.session.elicit("Are you sure you want to update your ticket to this new flight?",
                                           YesNoSchemaModel)
    if not user_response.answer:
        cursor.close()
        conn.close()
        return "Ticket update cancelled by user."

    # In a real application, you'd likely add additional checks here to enforce business logic,
    # like "does the new departure airport match the current ticket", etc.
    # While it's best to try to be *proactive* in 'type-hinting' policies to the LLM
    # it's inevitably going to get things wrong, so you **also** need to ensure your
    # API enforces valid behavior
    cursor.execute(
        "UPDATE ticket_flights SET flight_id = ? WHERE ticket_no = ?",
        (new_flight_id, ticket_no),
    )
    conn.commit()

    cursor.close()
    conn.close()
    return "Ticket successfully updated to new flight."


def get_passenger_id(context: Context) -> str:
    """Retrieve the passenger ID from the context or environment variable."""

    # Use the environment variable if it exists, this is useful for development/testing
    if os.environ.get("PASSENGER_ID"):
        return os.environ["PASSENGER_ID"]
    return context.request_context.passenger_id


@mcp.tool
def cancel_ticket(context: Context, ticket_no: str) -> str:
    """Cancel the user's ticket and remove it from the database."""
    passenger_id = get_passenger_id(context)
    conn = sqlite3.connect(db)
    cursor = conn.cursor()

    cursor.execute(
        "SELECT flight_id FROM ticket_flights WHERE ticket_no = ?", (ticket_no,)
    )
    existing_ticket = cursor.fetchone()
    if not existing_ticket:
        cursor.close()
        conn.close()
        return "No existing ticket found for the given ticket number."

    # Check the signed-in user actually has this ticket
    cursor.execute(
        "SELECT ticket_no FROM tickets WHERE ticket_no = ? AND passenger_id = ?",
        (ticket_no, passenger_id),
    )
    current_ticket = cursor.fetchone()
    if not current_ticket:
        cursor.close()
        conn.close()
        return f"Current signed-in passenger with ID {passenger_id} not the owner of ticket {ticket_no}"

    cursor.execute("DELETE FROM ticket_flights WHERE ticket_no = ?", (ticket_no,))
    conn.commit()

    cursor.close()
    conn.close()
    return "Ticket successfully cancelled."


if __name__ == "__main__":
    port = os.environ.get("FLIGHTS_PORT", "")
    if port:
        mcp.run(transport="streamable-http", port=int(port))
    else:
        mcp.run()
