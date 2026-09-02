import { AdminConsole } from "../components/admin-console";

export default function AdminHome() {
  return (
    <main className="console">
      <h1>Tale Role spectator</h1>
      <p>
        System admins join every room as <code>system_admin</code>. That seat is omitted from
        player roster, turn order, and Storyteller context.
      </p>
      <AdminConsole />
    </main>
  );
}
