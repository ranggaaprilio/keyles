import {
  NavigationMenu,
  NavigationMenuItem,
  NavigationMenuList,
  navigationMenuTriggerStyle,
} from "@/components/ui/navigation-menu";
import { Link } from "react-router-dom";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { isAuthenticated } from "@/services/api/auth";

export const Navbar = () => {
  const isLoggedIn = isAuthenticated();

  return (
    <div className="fixed top-0 left-0 right-0 z-50 bg-background/80 backdrop-blur-md border-b">
      <div className="container mx-auto px-4 h-16 flex items-center justify-between">
        <div className="font-bold text-xl">Keyles</div>

        <NavigationMenu>
          <NavigationMenuList>
            <NavigationMenuItem>
              <Link
                to="/#about"
                className={cn(navigationMenuTriggerStyle(), "cursor-pointer")}
              >
                About
              </Link>
            </NavigationMenuItem>
            <NavigationMenuItem>
              <Link
                to="/#pricing"
                className={cn(navigationMenuTriggerStyle(), "cursor-pointer")}
              >
                Pricing
              </Link>
            </NavigationMenuItem>
          </NavigationMenuList>
        </NavigationMenu>

        <div className="flex items-center gap-4">
          {isLoggedIn ? (
            <Button asChild>
              <Link to="/dashboard">Dashboard</Link>
            </Button>
          ) : (
            <Button asChild>
              <Link to="/login">Login</Link>
            </Button>
          )}
        </div>
      </div>
    </div>
  );
};
