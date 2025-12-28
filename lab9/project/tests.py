import pytest
from refactored_code import (
    NonNegativeFloatValidator,
    NonEmptyStringValidator,
    DevPayrollStrategy,
    ManagerPayrollStrategy,
    SalesPayrollStrategy,
    FixedPerformanceBonus,
    LevelBasedBonus,
    Employee,
    Developer,
    Manager,
    SalesPerson,
    MemoryStorage,
    Organization,
)


@pytest.fixture
def float_validator():
    return NonNegativeFloatValidator()


@pytest.fixture
def str_validator():
    return NonEmptyStringValidator()


def test_float_validator_valid(float_validator):
    assert float_validator.validate(100) == 100.0
    assert float_validator.validate(0) == 0.0


def test_float_validator_invalid(float_validator):
    with pytest.raises(ValueError):
        float_validator.validate(-5)


def test_string_validator_valid(str_validator):
    assert str_validator.validate("  Test  ") == "Test"


def test_string_validator_invalid(str_validator):
    with pytest.raises(ValueError):
        str_validator.validate("")


def test_dev_strategy():
    strat = DevPayrollStrategy()
    assert strat.compute(2000, "junior") == 2000
    assert strat.compute(2000, "senior") == 4000


def test_manager_strategy():
    strat = ManagerPayrollStrategy()
    assert strat.compute(7000, fixed_bonus=3000) == 10000


def test_sales_strategy():
    strat = SalesPayrollStrategy()
    assert strat.compute(3000, rate=0.15, sales_volume=4000) == 3600


def test_fixed_bonus():
    bonus = FixedPerformanceBonus()
    assert bonus.compute_bonus(5000) == 500.0


def test_level_bonus():
    bonus = LevelBasedBonus()
    assert bonus.compute_bonus(5000, "senior") == 1000.0


def test_developer_salary():
    dev = Developer("Test Dev", "DEV", 3000, "middle")
    assert dev.full_salary() == 4800.0  # 3000 * 1.5 + 3000 * 0.10


def test_manager_salary():
    mgr = Manager("Test Mgr", "MGMT", 6000, 1500)
    assert mgr.full_salary() == 8100.0  # 6000 + 1500 + 600


def test_sales_salary():
    sales = SalesPerson("Test Sales", "SALES", 2500, 0.12)
    sales.record_sale(8000)
    assert sales.full_salary() == 3460.0


def test_memory_storage():
    storage = MemoryStorage()
    emp = Employee("John", "IT", 4000)
    storage.save(emp)
    assert len(storage.list_all()) == 1
    assert storage.list_all()[0].emp_id == 1


def test_organization_flow():
    org = Organization("TestCo")
    dev = Developer("A", "DEV", 4000, "senior")
    mgr = Manager("B", "MGMT", 7000, 2000)
    org.add_employee(dev)
    org.add_employee(mgr)
    assert org.total_payroll() == 18100.0  # (4000*2 + 4000*0.2) + (7000+2000 + 700)


@pytest.mark.parametrize(
    "base,level,expected_mult,expected_bonus_rate",
    [(1000, "junior", 1.0, 0.05), (1000, "middle", 1.5, 0.10), (1000, "senior", 2.0, 0.20)],
)
def test_dev_parametrized(base, level, expected_mult, expected_bonus_rate):
    dev = Developer("P", "DEV", base, level)
    expected = base * expected_mult + base * expected_bonus_rate
    assert dev.full_salary() == expected


def test_zero_base_salary():
    dev = Developer("Zero", "DEV", 0, "junior")
    assert dev.full_salary() == 0.0
