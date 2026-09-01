export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body style={{ fontFamily: "Georgia, serif", background: "#120c0c", color: "#f3e7e7", margin: 0 }}>
        {children}
      </body>
    </html>
  );
}
