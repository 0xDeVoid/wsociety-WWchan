import { Link } from "react-router-dom";

export default function Logo() {
  return (
    <Link className="logo" to="/">
      <img src="/assets/logo.png" alt="WWchan" />
    </Link>
  );
}