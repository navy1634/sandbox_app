import Link from "next/link";
import Navbar, { type NavbarItem } from "../Navbar/components";
import styles from "./page.module.css";

export type HeaderProps = {
  navItems: NavbarItem[];
};

export default function Header({ navItems }: HeaderProps) {
  return (
    <div className={styles.headerShell}>
      <header className={styles.header}>
        <h1 className={styles.logo}>
          <Link href="/">iam</Link>
        </h1>
      </header>
      <Navbar items={navItems} />
    </div>
  );
}
