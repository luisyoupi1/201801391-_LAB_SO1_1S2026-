import json
import os
import random
from datetime import datetime, timezone
from pathlib import Path

from locust import HttpUser, between, task


DEFAULT_COUNTRIES = ["USA", "RUS", "CHN", "ESP", "GTM"]


def load_countries():
    catalog_path = Path(os.getenv("COUNTRIES_FILE", "/mnt/artifacts/countries.json"))
    if not catalog_path.is_file():
        return DEFAULT_COUNTRIES
    with catalog_path.open(encoding="utf-8") as catalog_file:
        catalog = json.load(catalog_file)
    countries = [entry["code"] for entry in catalog["countries"]]
    if sorted(countries) != sorted(DEFAULT_COUNTRIES):
        raise RuntimeError(f"catálogo de países inválido: {countries!r}")
    return countries


COUNTRIES = load_countries()


def report():
    return {
        "country": random.choice(COUNTRIES),
        "warplanes_in_air": random.randint(0, 50),
        "warships_in_water": random.randint(0, 30),
        "timestamp": datetime.now(timezone.utc).isoformat().replace("+00:00", "Z"),
    }


class MilitaryReportUser(HttpUser):
    wait_time = between(0.05, 0.4)

    @task(9)
    def grpc_flow(self):
        self.client.post("/grpc-201801391", json=report(), name="/grpc-201801391")

    @task(1)
    def dapr_flow(self):
        if os.getenv("ENABLE_DAPR", "false").lower() == "true":
            self.client.post("/dapr-201801391", json=report(), name="/dapr-201801391")
