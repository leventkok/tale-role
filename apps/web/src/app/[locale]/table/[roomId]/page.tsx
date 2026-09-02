import { setRequestLocale } from "next-intl/server";
import { TableClient } from "@/components/table-client";

export default async function TablePage({
  params,
}: {
  params: Promise<{ locale: string; roomId: string }>;
}) {
  const { locale, roomId } = await params;
  setRequestLocale(locale);
  return (
    <main>
      <TableClient roomId={roomId} />
    </main>
  );
}
