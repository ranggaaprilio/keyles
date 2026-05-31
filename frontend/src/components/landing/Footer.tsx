import { Link } from "react-router-dom";

export const Footer = () => {
  return (
    <footer className="bg-white">
      <div className="mx-auto max-w-[760px]">
        {/* Icon-label nav — Dell 1996 style */}
        <div className="border-t border-black">
          <div className="flex items-center justify-between px-4 py-4">
            {[
              { label: "PLATFORM", href: "/#how-it-works" },
              { label: "HOME", href: "/" },
              { label: "PLANS", href: "/#pricing" },
              { label: "SERVICE & SUPPORT", href: "/#about" },
            ].map((item) => (
              <Link
                key={item.label}
                to={item.href}
                className="flex flex-col items-center gap-1 px-2 text-center"
              >
                <span className="font-[Helvetica,Arial,system-ui,sans-serif] text-[11px] font-bold uppercase tracking-[1px] text-[#0000ee] underline">
                  {item.label}
                </span>
              </Link>
            ))}
          </div>
        </div>

        {/* Green connecting rule */}
        <div className="h-[1px] bg-[#8e8a25]" />

        {/* Copyright and legal */}
        <div className="px-4 py-4 text-center">
          <p className="font-['Times_New_Roman',Times,serif] text-xs text-gray-600">
            <a href="#" className="text-[#0000ee] underline">
              Copyright
            </a>{" "}
            &copy; {new Date().getFullYear()} Keyles. All rights reserved.{" "}
            <a href="#" className="text-[#0000ee] underline">
              (Terms of Use)
            </a>
          </p>
          <p className="mt-2 font-['Times_New_Roman',Times,serif] text-[11px] text-gray-400">
            This site is best viewed with browser versions 3.0 and higher.
          </p>
        </div>
      </div>
    </footer>
  );
};
