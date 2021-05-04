"""TESTS"""
load("transforms/t1/t1.star", "select_gpu_nodes", "lower_number_of_replicas")

test_data_path = "transforms/t1/testdata/"

def test_select_gpu_nodes():
    """test for select_gpu_nodes"""
    actual = yaml.load_file(test_data_path + "k8s-resources/demo-namespace.yaml")
    want = yaml.load_file(test_data_path + "want/demo-namespace.yaml")

    select_gpu_nodes(actual)

    if actual != want:
        fail("Test for select_gpu_nodes failed.\nExpected\n{}\nActual\n{}".format(want, actual))

def test_lower_number_of_replicas():
    """test for lower_number_of_replicas"""
    # sub test 1
    actual = yaml.load_file(test_data_path + "k8s-resources/javaspringapp-deployment.yaml")
    want = yaml.load_file(test_data_path + "want/javaspringapp-deployment.yaml")

    lower_number_of_replicas(actual)

    if actual != want:
        fail("Sub test 1 for lower_number_of_replicas failed.\nExpected\n{}\nActual\n{}".format(want, actual))

    # sub test 2
    actual = yaml.load_file(test_data_path + "k8s-resources/nginx-deployment.yaml")
    want = yaml.load_file(test_data_path + "want/nginx-deployment.yaml")

    lower_number_of_replicas(actual)

    if actual != want:
        fail("Sub test 2 for lower_number_of_replicas failed.\nExpected\n{}\nActual\n{}".format(want, actual))

def run_tests():
    test_select_gpu_nodes()
    test_lower_number_of_replicas()
