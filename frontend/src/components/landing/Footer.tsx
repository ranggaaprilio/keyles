export const Footer = () => {
  return (
    <footer className="bg-black py-16">
      <div className="mx-auto max-w-[1440px] px-5 md:px-8">
        <div className="mb-12 grid h-1 w-56 grid-cols-3">
          <span className="bg-[#0066b1]" />
          <span className="bg-[#1c69d4]" />
          <span className="bg-[#e22718]" />
        </div>

        <div className="grid gap-10 border-b border-[#3c3c3c] pb-12 md:grid-cols-4">
          <div>
            <p className="text-2xl font-bold uppercase tracking-[1.5px] text-white">
              Keyles
            </p>
            <p className="mt-4 max-w-xs text-sm font-light leading-6 text-[#bbbbbb]">
              Multi-tenant SSO for teams that need clear identity control.
            </p>
          </div>
          {[
            ["Platform", "OAuth clients", "User sessions", "Role mapping"],
            ["Operations", "Dashboard", "Tenant setup", "Audit trail"],
            ["Company", "Privacy policy", "Terms of service", "Contact"],
          ].map(([title, ...links]) => (
            <div key={title}>
              <p className="mb-4 text-sm font-bold uppercase tracking-[1.5px] text-white">
                {title}
              </p>
              <div className="space-y-3">
                {links.map((link) => (
                  <a
                    key={link}
                    href="#"
                    className="block text-sm font-light text-[#bbbbbb] transition-colors hover:text-white"
                  >
                    {link}
                  </a>
                ))}
              </div>
            </div>
          ))}
        </div>

        <div className="flex flex-col justify-between gap-4 pt-8 text-xs font-light tracking-[0.5px] text-[#7e7e7e] md:flex-row md:items-center">
          <p>&copy; {new Date().getFullYear()} Keyles. All rights reserved.</p>
          <p>Language: EN</p>
        </div>
      </div>
    </footer>
  );
};
